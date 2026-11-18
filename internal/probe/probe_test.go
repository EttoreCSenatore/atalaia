package probe

import (
	"net"
	"net/netip"
	"testing"

	"github.com/EttoreCSenatore/atalaia/internal/discovery"
	"github.com/EttoreCSenatore/atalaia/internal/fingerprint"
	"github.com/EttoreCSenatore/atalaia/internal/identity"
)

func host(ip, mac string) discovery.Host {
	m, _ := net.ParseMAC(mac)
	return discovery.Host{IP: netip.MustParseAddr(ip), MAC: m}
}

func TestCollapseKeepsEveryAddressOfOneMAC(t *testing.T) {
	hosts := []discovery.Host{
		host("192.168.1.1", "aa:00:00:00:00:01"),
		host("192.168.1.5", "aa:00:00:00:00:02"),
		host("192.168.1.6", "aa:00:00:00:00:02"),
		host("192.168.1.7", "aa:00:00:00:00:02"),
	}
	obs := Collapse(hosts, netip.MustParseAddr("192.168.1.1"), map[string]string{"192.168.1.6": "evil\x1b[31m.lan"}, nil, identity.Default())
	if len(obs) != 2 {
		t.Fatalf("want 2 observations, got %d", len(obs))
	}
	if !obs[0].Gateway || len(obs[0].IPs) != 1 {
		t.Fatalf("gateway: %+v", obs[0])
	}
	if obs[1].Gateway || len(obs[1].IPs) != 3 || obs[1].IP.String() != "192.168.1.5" {
		t.Fatalf("multi: %+v", obs[1])
	}
	if obs[1].Name != "evil[31m.lan" && obs[1].Name != "evil.lan" && obs[1].Name == "" {
		t.Fatalf("name not carried: %q", obs[1].Name)
	}
}

func TestCollapseShowsTheGatewayAddressWhenAMACClaimsIt(t *testing.T) {
	hosts := []discovery.Host{
		host("192.168.1.1", "aa:00:00:00:00:09"),
		host("192.168.1.2", "aa:00:00:00:00:09"),
	}
	obs := Collapse(hosts, netip.MustParseAddr("192.168.1.2"), nil, nil, identity.Default())
	if len(obs) != 1 || !obs[0].Gateway || obs[0].IP.String() != "192.168.1.2" {
		t.Fatalf("%+v", obs)
	}
}

func TestAppendSelfAddsOnceAndKeepsOrder(t *testing.T) {
	hosts := []discovery.Host{host("192.168.1.1", "aa:00:00:00:00:01"), host("192.168.1.30", "aa:00:00:00:00:03")}
	self := []discovery.Host{host("192.168.1.20", "aa:00:00:00:00:02"), host("192.168.1.30", "aa:00:00:00:00:03")}
	got := appendSelf(hosts, self)
	if len(got) != 3 || got[1].IP.String() != "192.168.1.20" {
		t.Fatalf("%+v", got)
	}
}

func TestCollapseMergesWhatTheDeviceAnnounced(t *testing.T) {
	hosts := []discovery.Host{host("192.168.1.40", "da:00:00:00:00:01"), host("192.168.1.41", "a8:00:00:00:00:02")}
	learned := map[netip.Addr]fingerprint.Result{
		netip.MustParseAddr("192.168.1.40"): {Services: []string{"_googlecast._tcp"}, Model: "Chromecast", Friendly: "Sala"},
		netip.MustParseAddr("192.168.1.41"): {HasNetBIOS: true, NetBIOS: fingerprint.NetBIOSInfo{Name: "DESKTOP-1"}, Manufacturer: "Acme PC"},
	}
	obs := Collapse(hosts, netip.Addr{}, nil, learned, identity.Default())
	if obs[0].Kind != "tv" || obs[0].Model != "Chromecast" || obs[0].Name != "Sala" || len(obs[0].Services) != 1 {
		t.Fatalf("%+v", obs[0])
	}
	if obs[1].Kind != "computer" || obs[1].Name != "DESKTOP-1" || obs[1].Vendor == "" {
		t.Fatalf("%+v", obs[1])
	}
	// a name found by mDNS/reverse DNS is not replaced by a weaker source
	obs = Collapse(hosts[:1], netip.Addr{}, map[string]string{"192.168.1.40": "living-room-tv"}, learned, identity.Default())
	if obs[0].Name != "living-room-tv" {
		t.Fatalf("%q", obs[0].Name)
	}
}
