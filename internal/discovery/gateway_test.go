package discovery

import (
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

func TestParseGateway(t *testing.T) {
	const table = `Iface	Destination	Gateway 	Flags	RefCnt	Use	Metric	Mask		MTU	Window	IRTT
eth0	00000000	0101A8C0	0003	0	0	100	00000000	0	0	0
eth0	0001A8C0	00000000	0001	0	0	100	00FFFFFF	0	0	0
`
	gw, ok := ParseGateway(strings.NewReader(table))
	if !ok || gw.String() != "192.168.1.1" {
		t.Fatalf("got %v %v", gw, ok)
	}
	for name, in := range map[string]string{
		"no default": "Iface\tDestination\tGateway\tFlags\tRefCnt\tUse\tMetric\tMask\nlo\t0001A8C0\t00000000\t0001\t0\t0\t0\t00FFFFFF\n",
		"route down": "h\neth0\t00000000\t0101A8C0\t0002\t0\t0\t0\t00000000\n",
		"garbage":    "h\nx y z\n",
		"empty":      "",
	} {
		if gw, ok := ParseGateway(strings.NewReader(in)); ok {
			t.Errorf("%s: unexpected %v", name, gw)
		}
	}
}

func TestSelfHostsOnlyTheInterfaceOnTheGatewayNetwork(t *testing.T) {
	mk := func(name string, flags net.Flags, mac string) net.Interface {
		hw, _ := net.ParseMAC(mac)
		return net.Interface{Name: name, Flags: flags, HardwareAddr: hw}
	}
	ifs := []net.Interface{
		mk("lo", net.FlagUp|net.FlagLoopback, ""),
		mk("eth0", net.FlagUp, "aa:00:00:00:00:01"),
		mk("docker0", net.FlagUp, "aa:00:00:00:00:02"),
		mk("wlan0", 0, "aa:00:00:00:00:03"), // down
	}
	addrs := map[string]string{"eth0": "192.168.1.20/24", "docker0": "172.17.0.1/16", "wlan0": "192.168.1.21/24", "lo": "127.0.0.1/8"}
	get := func(i *net.Interface) ([]net.Addr, error) {
		ip, n, _ := net.ParseCIDR(addrs[i.Name])
		n.IP = ip
		return []net.Addr{n}, nil
	}
	sc := scope.New([]netip.Prefix{netip.MustParsePrefix("192.168.1.0/24"), netip.MustParsePrefix("172.17.0.0/16")})
	got := selfHosts(netip.MustParseAddr("192.168.1.1"), sc, ifs, get)
	if len(got) != 1 || got[0].Iface != "eth0" || got[0].IP.String() != "192.168.1.20" {
		t.Fatalf("%+v", got)
	}
	if got := selfHosts(netip.Addr{}, sc, ifs, get); len(got) != 0 {
		t.Fatalf("no gateway, no self: %+v", got)
	}
}
