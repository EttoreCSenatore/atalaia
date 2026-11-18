package scope

import (
	"net"
	"net/netip"
	"testing"
)

func pfx(s string) netip.Prefix { return netip.MustParsePrefix(s) }
func ip(s string) netip.Addr    { return netip.MustParseAddr(s) }

func TestNewKeepsOnlyPrivateIPv4(t *testing.T) {
	s := New([]netip.Prefix{
		pfx("192.168.1.20/24"),
		pfx("8.8.8.8/24"),           // public: dropped
		pfx("fe80::1/64"),           // IPv6: dropped
		pfx("127.0.0.1/8"),          // loopback: dropped
		pfx("169.254.3.4/16"),       // link-local: dropped
		pfx("192.168.1.99/24"),      // duplicate network
		pfx("10.1.2.3/8"),           // too large: clamped to /16
		pfx("::ffff:172.16.0.5/24"), // IPv4-mapped: unmapped and kept
	})
	want := []string{"10.1.0.0/16", "172.16.0.0/24", "192.168.1.0/24"}
	got := s.Prefixes()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Errorf("prefix %d = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestAllows(t *testing.T) {
	s := New([]netip.Prefix{pfx("192.168.1.20/24"), pfx("10.0.0.5/31")})
	cases := []struct {
		addr string
		want bool
	}{
		{"192.168.1.1", true},
		{"192.168.1.254", true},
		{"192.168.1.0", false},   // network address
		{"192.168.1.255", false}, // broadcast
		{"192.168.2.1", false},   // other network
		{"8.8.8.8", false},       // internet
		{"10.0.0.4", true},       // /31 has no network/broadcast
		{"10.0.0.5", true},
		{"10.0.0.6", false},
		{"::ffff:192.168.1.7", true},
		{"fe80::1", false},
	}
	for _, c := range cases {
		if got := s.Allows(ip(c.addr)); got != c.want {
			t.Errorf("Allows(%s) = %v, want %v", c.addr, got, c.want)
		}
	}
}

func TestEmptyScopeAllowsNothing(t *testing.T) {
	if New(nil).Allows(ip("192.168.1.1")) {
		t.Fatal("empty scope must allow nothing")
	}
}

func TestFromAddrs(t *testing.T) {
	_, n, _ := net.ParseCIDR("192.168.5.0/24")
	n.IP = net.ParseIP("192.168.5.9")
	s := FromAddrs([]net.Addr{n, &net.IPAddr{IP: net.ParseIP("10.0.0.1")}})
	if got := s.Prefixes(); len(got) != 1 || got[0].String() != "192.168.5.0/24" {
		t.Fatalf("got %v", got)
	}
}

func TestSweepTargets(t *testing.T) {
	s := New([]netip.Prefix{pfx("192.168.1.20/24")})
	got := s.SweepTargets()
	if len(got) != 253 { // 254 hosts minus the host itself
		t.Fatalf("got %d targets, want 253", len(got))
	}
	for _, a := range got {
		if a == ip("192.168.1.20") || !s.Allows(a) {
			t.Fatalf("bad target %s", a)
		}
	}
	big := New([]netip.Prefix{pfx("10.1.2.3/16")})
	targets := big.SweepTargets()
	if len(targets) != 253 || targets[0] != ip("10.1.2.1") || targets[252] != ip("10.1.2.254") {
		t.Fatalf("large network must sweep only the local /24, got %d (%v..)", len(targets), targets[:1])
	}
}
