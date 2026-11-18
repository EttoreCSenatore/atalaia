package dhcpprobe

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"testing"
	"time"
)

func offer(xid uint32, server, offered, router string, dns []string, typ byte) []byte {
	p := make([]byte, 240)
	p[0] = 2
	binary.BigEndian.PutUint32(p[4:8], xid)
	copy(p[16:20], net.ParseIP(offered).To4())
	copy(p[236:240], magicCookie)
	p = append(p, optMessageType, 1, typ)
	p = append(p, optServerID, 4)
	p = append(p, net.ParseIP(server).To4()...)
	p = append(p, optRouter, 4)
	p = append(p, net.ParseIP(router).To4()...)
	p = append(p, optDNS, byte(4*len(dns)))
	for _, d := range dns {
		p = append(p, net.ParseIP(d).To4()...)
	}
	p = append(p, optLease, 4, 0, 0, 0x0e, 0x10) // 3600 s
	return append(p, optEnd)
}

func TestBuildDiscoverIsAValidBroadcastDiscover(t *testing.T) {
	mac := RandomMAC()
	if mac[0]&0x02 == 0 || mac[0]&0x01 != 0 {
		t.Fatalf("not a locally administered unicast MAC: %s", mac)
	}
	p := BuildDiscover(0xdeadbeef, mac)
	if p[0] != 1 || binary.BigEndian.Uint32(p[4:8]) != 0xdeadbeef || binary.BigEndian.Uint16(p[10:12]) != 0x8000 {
		t.Fatalf("header %v", p[:12])
	}
	if string(p[28:34]) != string(mac) || string(p[236:240]) != string(magicCookie) {
		t.Fatal("client MAC or cookie wrong")
	}
	// message type option says DISCOVER, and the packet ends with the end option
	if p[240] != optMessageType || p[242] != typeDiscover || p[len(p)-1] != optEnd {
		t.Fatalf("options %v", p[240:])
	}
}

func TestParseOffer(t *testing.T) {
	p := offer(7, "192.168.1.1", "192.168.1.50", "192.168.1.1", []string{"1.1.1.1", "8.8.8.8"}, typeOffer)
	o, ok := ParseOffer(p, 7, netip.MustParseAddr("192.168.1.1"))
	if !ok || o.Server.String() != "192.168.1.1" || o.Offered.String() != "192.168.1.50" || o.Router.String() != "192.168.1.1" ||
		len(o.DNS) != 2 || o.DNS[1].String() != "8.8.8.8" || o.LeaseS != 3600 {
		t.Fatalf("%+v %v", o, ok)
	}
	for name, bad := range map[string][]byte{
		"other transaction":  offer(8, "192.168.1.1", "192.168.1.50", "192.168.1.1", nil, typeOffer),
		"an ACK, not offer":  offer(7, "192.168.1.1", "192.168.1.50", "192.168.1.1", nil, 5),
		"too short":          p[:100],
		"truncated option":   p[:len(p)-8],
		"request, not reply": append([]byte{1}, p[1:]...),
	} {
		if _, ok := ParseOffer(bad, 7, netip.Addr{}); ok {
			t.Errorf("%s was accepted", name)
		}
	}
	// without option 54 the sender is the server
	q := offer(7, "0.0.0.0", "192.168.1.50", "192.168.1.1", nil, typeOffer)
	if o, ok := ParseOffer(q, 7, netip.MustParseAddr("192.168.1.77")); !ok || o.Server.IsUnspecified() && o.From.String() != "192.168.1.77" {
		t.Fatalf("%+v", o)
	}
}

// fakeServers answers a discover with offers from several servers on loopback.
func TestProbeCollectsAndDeduplicatesOffers(t *testing.T) {
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	client, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	go func() {
		buf := make([]byte, 1500)
		n, from, err := server.ReadFromUDP(buf)
		if err != nil || n < 241 {
			return
		}
		xid := binary.BigEndian.Uint32(buf[4:8])
		for _, s := range []string{"192.168.1.1", "192.168.1.200", "192.168.1.1"} { // the first twice
			_, _ = server.WriteToUDP(offer(xid, s, "192.168.1.50", s, []string{"9.9.9.9"}, typeOffer), from)
		}
		_, _ = server.WriteToUDP(offer(xid+1, "10.9.9.9", "192.168.1.51", "10.9.9.9", nil, typeOffer), from) // other transaction
	}()
	got, err := Probe(context.Background(), client, server.LocalAddr().(*net.UDPAddr), 400*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Server.String() != "192.168.1.1" || got[1].Server.String() != "192.168.1.200" {
		t.Fatalf("%+v", got)
	}
}

func TestProbeWithNoAnswerReturnsNothing(t *testing.T) {
	server, _ := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	defer server.Close()
	client, _ := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	defer client.Close()
	got, err := Probe(context.Background(), client, server.LocalAddr().(*net.UDPAddr), 150*time.Millisecond)
	if err != nil || len(got) != 0 {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestIsPermission(t *testing.T) {
	_, err := Listen(68) // fails without privileges (as in CI); succeeds as root
	if err != nil && !IsPermission(err) {
		t.Logf("port 68 unavailable for another reason: %v", err)
	}
}
