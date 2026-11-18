package fingerprint

import (
	"context"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// fakeMDNS answers like an Apple TV / Chromecast on a loopback port.
func fakeMDNS(t *testing.T, services []string, txt map[string][]string) int {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 1500)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			q := new(dns.Msg)
			if q.Unpack(buf[:n]) != nil || len(q.Question) != 1 {
				continue
			}
			r := new(dns.Msg)
			r.SetReply(q)
			name := q.Question[0].Name
			switch {
			case name == "_services._dns-sd._udp.local.":
				for _, s := range services {
					r.Answer = append(r.Answer, &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 10}, Ptr: s + ".local."})
				}
				r.Answer = append(r.Answer, &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 10}, Ptr: "evil\x00name.local."})
			case q.Question[0].Qtype == dns.TypePTR:
				r.Answer = append(r.Answer, &dns.PTR{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 10}, Ptr: "Living Room." + name})
			case q.Question[0].Qtype == dns.TypeTXT:
				for svc, kv := range txt {
					if len(name) > len(svc) && name[len(name)-len(svc)-7:] == svc+".local." {
						r.Answer = append(r.Answer, &dns.TXT{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 10}, Txt: kv})
					}
				}
			}
			if len(r.Answer) > 0 {
				b, _ := r.Pack()
				_, _ = pc.WriteTo(b, from)
			}
		}
	}()
	return pc.LocalAddr().(*net.UDPAddr).Port
}

func TestMDNSServicesAndModel(t *testing.T) {
	port := fakeMDNS(t, []string{"_googlecast._tcp", "_airplay._tcp", "_http._tcp", "_http._tcp"},
		map[string][]string{"_googlecast._tcp": {"id=abc", "md=Chromecast Ultra\x1b[31m", "fn=Sala"}})
	info := MDNS{Port: port, Timeout: time.Second}.Info(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if len(info.Services) != 3 || info.Services[0] != "_googlecast._tcp" {
		t.Fatalf("services (deduplicated, malformed ones dropped): %+v", info.Services)
	}
	if !strings.HasPrefix(info.Model, "Chromecast Ultra") {
		t.Fatalf("model: %q", info.Model)
	}
	for _, r := range info.Model {
		if r < 0x20 {
			t.Fatalf("control character kept: %q", info.Model)
		}
	}
}

func TestMDNSSilentDevice(t *testing.T) {
	pc, _ := net.ListenPacket("udp", "127.0.0.1:0")
	defer pc.Close()
	port := pc.LocalAddr().(*net.UDPAddr).Port
	info := MDNS{Port: port, Timeout: 100 * time.Millisecond}.Info(context.Background(), netip.MustParseAddr("127.0.0.1"))
	if len(info.Services) != 0 || info.Model != "" {
		t.Fatalf("%+v", info)
	}
}
