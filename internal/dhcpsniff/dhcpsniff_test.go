package dhcpsniff

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func request(mac net.HardwareAddr, typ byte, hostname, vendor string) []byte {
	p := make([]byte, 240)
	p[0], p[1], p[2] = 1, 1, 6
	copy(p[28:34], mac)
	copy(p[236:240], cookie)
	p = append(p, 53, 1, typ)
	if hostname != "" {
		p = append(p, 12, byte(len(hostname)))
		p = append(p, hostname...)
	}
	if vendor != "" {
		p = append(p, 60, byte(len(vendor)))
		p = append(p, vendor...)
	}
	return append(p, 255)
}

func TestParse(t *testing.T) {
	mac, _ := net.ParseMAC("da:44:55:00:00:14")
	c, ok := Parse(request(mac, 3, "Marias-iPhone", "android-dhcp-14"))
	if !ok || c.MAC != "da:44:55:00:00:14" || c.Hostname != "Marias-iPhone" || c.VendorClass != "android-dhcp-14" {
		t.Fatalf("%+v %v", c, ok)
	}
	// hostile text is cleaned
	c, ok = Parse(request(mac, 1, "evil\x1b[31m\r\nname", ""))
	if !ok || c.Hostname == "" {
		t.Fatalf("%+v", c)
	}
	for _, r := range c.Hostname {
		if r < 0x20 {
			t.Fatalf("control character kept: %q", c.Hostname)
		}
	}
	// a server's OFFER, a malformed packet and a zero MAC are ignored
	if _, ok := Parse(request(mac, 2, "x", "")); ok {
		t.Error("an OFFER type was accepted")
	}
	offer := request(mac, 3, "x", "")
	offer[0] = 2
	if _, ok := Parse(offer); ok {
		t.Error("a BOOTREPLY was accepted")
	}
	if _, ok := Parse(request(mac, 3, "x", "")[:200]); ok {
		t.Error("a short packet was accepted")
	}
	trunc := request(mac, 3, "hostname", "")
	if _, ok := Parse(trunc[:len(trunc)-5]); ok {
		t.Error("a truncated option was accepted")
	}
	if _, ok := Parse(request(net.HardwareAddr{0, 0, 0, 0, 0, 0}, 3, "x", "")); ok {
		t.Error("a zero MAC was accepted")
	}
}

func TestKindFromVendorClass(t *testing.T) {
	for in, want := range map[string]string{"android-dhcp-13": "phone", "MSFT 5.0": "computer", "udhcp 1.30": "", "": ""} {
		if got := KindFromVendorClass(in); got != want {
			t.Errorf("%q: %q want %q", in, got, want)
		}
	}
}

func TestSnifferRunsWhileEnabledAndReportsListenErrors(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var got []Client
	var on = true
	var failFirst = true
	s := &Sniffer{
		Listen: func() (net.PacketConn, error) {
			mu.Lock()
			defer mu.Unlock()
			if failFirst {
				failFirst = false
				return nil, net.ErrClosed
			}
			return pc, nil
		},
		Handle: func(_ context.Context, c Client) { mu.Lock(); got = append(got, c); mu.Unlock() },
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.Run(ctx, func() bool { mu.Lock(); defer mu.Unlock(); return on }, 30*time.Millisecond, func(error) string { return "needs port 67" })
		close(done)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if a, _ := s.Status(); a {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if a, e := s.Status(); !a || e != "" {
		t.Fatalf("not active after retrying: %v %q", a, e)
	}
	mac, _ := net.ParseMAC("da:44:55:00:00:14")
	cl, _ := net.Dial("udp", pc.LocalAddr().String())
	_, _ = cl.Write(request(mac, 3, "pixel-8", "android-dhcp-14"))
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(got)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mu.Lock()
	if len(got) != 1 || got[0].Hostname != "pixel-8" {
		t.Fatalf("%+v", got)
	}
	on = false
	mu.Unlock()
	for time.Now().Before(deadline) {
		if a, _ := s.Status(); !a {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if a, _ := s.Status(); a {
		t.Fatal("still listening after being disabled")
	}
	cancel()
	<-done
}
