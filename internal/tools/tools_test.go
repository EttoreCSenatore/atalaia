package tools

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"
)

func TestMagicPacket(t *testing.T) {
	mac, _ := net.ParseMAC("aa:bb:cc:00:11:22")
	p, err := MagicPacket(mac)
	if err != nil || len(p) != 102 {
		t.Fatalf("%d %v", len(p), err)
	}
	if !bytes.Equal(p[:6], bytes.Repeat([]byte{0xff}, 6)) || !bytes.Equal(p[96:], mac) || !bytes.Equal(p[6:12], mac) {
		t.Fatal("wrong layout")
	}
	if _, err := MagicPacket(net.HardwareAddr{1, 2}); err == nil {
		t.Fatal("short MAC accepted")
	}
}

func TestBroadcast(t *testing.T) {
	for in, want := range map[string]string{"192.168.1.0/24": "192.168.1.255", "10.0.0.0/16": "10.0.255.255", "172.16.4.9/20": "172.16.15.255", "192.168.1.128/25": "192.168.1.255"} {
		if got := Broadcast(netip.MustParsePrefix(in)); got.String() != want {
			t.Errorf("%s: %s want %s", in, got, want)
		}
	}
}

func TestResolveIPv4Refusals(t *testing.T) {
	ctx := context.Background()
	for _, bad := range []string{"", "224.0.0.1", "0.0.0.0", "255.255.255.255", "::1", "2001:db8::1"} {
		if _, err := ResolveIPv4(ctx, bad); err == nil {
			t.Errorf("%q should be refused", bad)
		}
	}
	if ip, err := ResolveIPv4(ctx, "192.168.1.1"); err != nil || ip.String() != "192.168.1.1" {
		t.Fatalf("%v %v", ip, err)
	}
}

func TestPortScanReportsOnlyCompletedHandshakes(t *testing.T) {
	var asked []string
	dial := func(ctx context.Context, addr string) (net.Conn, error) {
		asked = append(asked, addr)
		if addr == "192.168.1.9:22" || addr == "192.168.1.9:8080" {
			c, _ := net.Pipe()
			return c, nil
		}
		return nil, errors.New("refused")
	}
	got := PortScan(context.Background(), netip.MustParseAddr("192.168.1.9"), dial, 1, time.Second)
	if len(got) != 2 || got[0].Port != 22 || got[0].Service != "ssh" || got[1].Port != 8080 {
		t.Fatalf("%+v", got)
	}
	if len(asked) != len(CommonPorts) {
		t.Fatalf("asked %d ports, want %d", len(asked), len(CommonPorts))
	}
}

func TestPortScanStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	n := 0
	got := PortScan(ctx, netip.MustParseAddr("192.168.1.9"), func(context.Context, string) (net.Conn, error) { n++; return nil, errors.New("x") }, 1, time.Second)
	if len(got) != 0 || n > 1 {
		t.Fatalf("scanned after cancel: %d", n)
	}
}

func TestDNSLookupValidationAndCleaning(t *testing.T) {
	ctx := context.Background()
	if _, err := DNSLookup(ctx, nil, "", "A"); err == nil {
		t.Error("empty name")
	}
	if _, err := DNSLookup(ctx, nil, "x.test", "ANY"); err == nil {
		t.Error("unsupported type")
	}
	if _, err := DNSLookup(ctx, nil, "not-an-ip", "PTR"); err == nil {
		t.Error("PTR needs an IP")
	}
	if got := clean("a\r\nb\x1b[0m", 300); got != "a  b [0m" {
		t.Errorf("clean: %q", got)
	}
}
