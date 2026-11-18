package discovery

import (
	"strings"
	"testing"
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
