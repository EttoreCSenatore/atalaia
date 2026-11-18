package discovery

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

const arpSample = `IP address       HW type     Flags       HW address            Mask     Device
192.168.1.1      0x1         0x2         a8:bb:cc:00:00:01     *        eth0
192.168.1.30     0x1         0x0         00:00:00:00:00:00     *        eth0
192.168.1.40     0x1         0x2         00:00:00:00:00:00     *        eth0
192.168.1.50     0x1         0x2         a8:bb:cc:00:00:05     *        eth0
8.8.8.8          0x1         0x2         a8:bb:cc:00:00:08     *        eth0
not-an-ip        0x1         0x2         a8:bb:cc:00:00:09     *        eth0
short line
`

func TestParseARP(t *testing.T) {
	hosts, err := ParseARP(strings.NewReader(arpSample))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, h := range hosts {
		got = append(got, h.IP.String()+" "+h.MAC.String()+" "+h.Iface)
	}
	want := []string{
		"192.168.1.1 a8:bb:cc:00:00:01 eth0",
		"192.168.1.50 a8:bb:cc:00:00:05 eth0",
		"8.8.8.8 a8:bb:cc:00:00:08 eth0",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestScanNudgesOnlyScopeAndFiltersResults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "arp")
	if err := os.WriteFile(path, []byte(arpSample), 0o600); err != nil {
		t.Fatal(err)
	}
	sc := scope.New([]netip.Prefix{netip.MustParsePrefix("192.168.1.20/24")})

	var mu sync.Mutex
	nudged := map[netip.Addr]bool{}
	s := &Scanner{
		ARPPath: path,
		Wait:    time.Millisecond,
		Nudge: func(_ context.Context, ip netip.Addr) {
			mu.Lock()
			nudged[ip] = true
			mu.Unlock()
		},
	}
	hosts, err := s.Scan(context.Background(), sc)
	if err != nil {
		t.Fatal(err)
	}
	if len(nudged) != 253 {
		t.Fatalf("nudged %d addresses, want 253", len(nudged))
	}
	for ip := range nudged {
		if !sc.Allows(ip) {
			t.Fatalf("probed %s outside the scope", ip)
		}
	}
	if len(hosts) != 2 || hosts[0].IP.String() != "192.168.1.1" || hosts[1].IP.String() != "192.168.1.50" {
		t.Fatalf("hosts = %+v (8.8.8.8 is out of scope and must be dropped)", hosts)
	}
}

func TestScanUnsupportedWithoutTable(t *testing.T) {
	s := &Scanner{ARPPath: filepath.Join(t.TempDir(), "missing")}
	_, err := s.Scan(context.Background(), scope.New(nil))
	if err == nil || !strings.Contains(err.Error(), "Linux only") {
		t.Fatalf("err = %v", err)
	}
}

func TestScanStopsOnCancel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arp")
	_ = os.WriteFile(path, []byte(arpSample), 0o600)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := &Scanner{ARPPath: path, Wait: time.Hour, Nudge: func(context.Context, netip.Addr) {}}
	sc := scope.New([]netip.Prefix{netip.MustParsePrefix("192.168.1.20/24")})
	if _, err := s.Scan(ctx, sc); err == nil {
		t.Fatal("expected context error")
	}
}
