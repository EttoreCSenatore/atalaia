package names

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// fakeMDNS answers PTR queries on a loopback UDP port with the given name,
// preceded by an unrelated answer to check that it is ignored.
func fakeMDNS(t *testing.T, ptrName string) int {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 1500)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			q := new(dns.Msg)
			if q.Unpack(buf[:n]) != nil || len(q.Question) == 0 {
				continue
			}
			r := new(dns.Msg)
			r.SetReply(q)
			r.Answer = append(r.Answer,
				&dns.PTR{Hdr: dns.RR_Header{Name: "other.in-addr.arpa.", Rrtype: dns.TypePTR, Class: dns.ClassINET}, Ptr: "wrong.local."},
				&dns.PTR{Hdr: dns.RR_Header{Name: q.Question[0].Name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 120}, Ptr: ptrName},
			)
			if b, err := r.Pack(); err == nil {
				_, _ = pc.WriteTo(b, addr)
			}
		}
	}()
	return pc.LocalAddr().(*net.UDPAddr).Port
}

func TestMDNSReverseLookup(t *testing.T) {
	port := fakeMDNS(t, "Marias-iPhone.local.")
	got := MDNS{Port: port, Timeout: time.Second}.Name(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if got != "Marias-iPhone.local" {
		t.Fatalf("got %q", got)
	}
}

func TestMDNSNoAnswerTimesOut(t *testing.T) {
	pc, _ := net.ListenPacket("udp", "127.0.0.1:0") // listens, never answers
	defer pc.Close()
	port := pc.LocalAddr().(*net.UDPAddr).Port
	start := time.Now()
	got := MDNS{Port: port, Timeout: 100 * time.Millisecond}.Name(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if got != "" || time.Since(start) > time.Second {
		t.Fatalf("got %q after %v", got, time.Since(start))
	}
}

type fixed map[string]string

func (f fixed) Name(_ context.Context, ip netip.Addr) string { return f[ip.String()] }

func TestChainAndResolveAll(t *testing.T) {
	chain := Chain{fixed{"10.0.0.1": "first"}, fixed{"10.0.0.1": "second", "10.0.0.2": "other"}}
	ips := []netip.Addr{netip.MustParseAddr("10.0.0.1"), netip.MustParseAddr("10.0.0.2"), netip.MustParseAddr("10.0.0.3")}
	got := ResolveAll(context.Background(), chain, ips, 2)
	if len(got) != 2 || got[ips[0]] != "first" || got[ips[1]] != "other" {
		t.Fatalf("got %v", got)
	}
}
