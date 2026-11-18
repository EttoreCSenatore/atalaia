package probe

import (
	"net"
	"net/netip"
	"testing"

	"github.com/EttoreCSenatore/atalaia/internal/discovery"
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
	obs := Collapse(hosts, netip.MustParseAddr("192.168.1.1"), map[string]string{"192.168.1.6": "evil\x1b[31m.lan"}, identity.Default())
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
	obs := Collapse(hosts, netip.MustParseAddr("192.168.1.2"), nil, identity.Default())
	if len(obs) != 1 || !obs[0].Gateway || obs[0].IP.String() != "192.168.1.2" {
		t.Fatalf("%+v", obs)
	}
}
