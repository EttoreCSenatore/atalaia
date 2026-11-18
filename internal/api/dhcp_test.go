package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/dhcpprobe"
)

func offerFor(ip string) dhcpprobe.Offer {
	a := netip.MustParseAddr(ip)
	return dhcpprobe.Offer{Server: a, From: a, Router: a, DNS: []netip.Addr{netip.MustParseAddr("6.6.6.6")}}
}

func TestDHCPCheckFlow(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	answers := [][]dhcpprobe.Offer{{offerFor("192.168.1.1")}, {offerFor("192.168.1.1"), offerFor("192.168.1.77")}}
	i := 0
	f.srv.DHCP.Probe = func(context.Context) ([]dhcpprobe.Offer, error) { a := answers[i]; i++; return a, nil }

	var v dhcpDTO
	_ = json.Unmarshal(f.do("GET", "/api/dhcp", nil, c).Body.Bytes(), &v)
	if v.Enabled || len(v.Servers) != 0 || v.Last != nil {
		t.Fatalf("off by default: %+v", v)
	}
	if w := f.do("POST", "/api/dhcp/check", map[string]bool{"confirm": false}, c); w.Code != 400 {
		t.Fatalf("needs confirmation: %d", w.Code)
	}
	if w := f.do("POST", "/api/dhcp/check", map[string]bool{"confirm": true}, c); w.Code != 200 {
		t.Fatalf("check: %d %s", w.Code, w.Body)
	}
	w := f.do("POST", "/api/dhcp/check", map[string]bool{"confirm": true}, c)
	_ = json.Unmarshal(w.Body.Bytes(), &v)
	if len(v.Servers) != 2 || v.Last == nil || len(v.Last.New) != 1 || v.Last.New[0] != "192.168.1.77" {
		t.Fatalf("rogue not reported: %+v", v)
	}
	if v.Servers[0].Router != "192.168.1.1" || len(v.Servers[0].DNS) != 1 {
		t.Fatalf("detail: %+v", v.Servers[0])
	}
	var a []auditDTO
	_ = json.Unmarshal(f.do("GET", "/api/audit", nil, c).Body.Bytes(), &a)
	if len(a) < 2 || a[0].Action != "dhcp" {
		t.Fatalf("audit: %+v", a)
	}
	// enable, then reset
	if w := f.do("PUT", "/api/dhcp", map[string]bool{"enabled": true}, c); w.Code != 200 {
		t.Fatalf("enable: %d", w.Code)
	}
	if !f.srv.DHCP.Enabled(context.Background()) {
		t.Fatal("not enabled")
	}
	if w := f.do("POST", "/api/dhcp/reset", map[string]bool{"confirm": true}, c); w.Code != 200 {
		t.Fatalf("reset: %d", w.Code)
	}
	if l, _ := f.st.DHCPServers(context.Background()); len(l) != 0 {
		t.Fatalf("not forgotten: %+v", l)
	}
	if w := f.do("GET", "/api/dhcp", nil, nil); w.Code != 401 {
		t.Fatalf("anonymous: %d", w.Code)
	}
}

func TestDHCPProbeFailureIsExplainedAndNotABaseline(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.srv.DHCP.Probe = func(context.Context) ([]dhcpprobe.Offer, error) {
		return nil, errors.New("binding UDP port 68 needs root or CAP_NET_BIND_SERVICE")
	}
	w := f.do("POST", "/api/dhcp/check", map[string]bool{"confirm": true}, c)
	if w.Code != 502 || !strings.Contains(w.Body.String(), "CAP_NET_BIND_SERVICE") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	if l, _ := f.st.DHCPServers(context.Background()); len(l) != 0 {
		t.Fatal("a failed probe recorded servers")
	}
}

func TestDHCPIsDisabledInObserveOnlyMode(t *testing.T) {
	f := newFixture(t)
	f.srv.NoActiveTools = true
	f.h = f.srv.Handler()
	c := f.setup()
	f.srv.DHCP.Probe = func(context.Context) ([]dhcpprobe.Offer, error) {
		t.Fatal("probed in observe-only mode")
		return nil, nil
	}
	for _, req := range [][3]string{{"PUT", "/api/dhcp", `{"enabled":true}`}, {"POST", "/api/dhcp/check", `{"confirm":true}`}, {"POST", "/api/dhcp/reset", `{"confirm":true}`}} {
		var body any
		_ = json.Unmarshal([]byte(req[2]), &body)
		if w := f.do(req[0], req[1], body, c); w.Code != 403 {
			t.Errorf("%s %s: %d", req[0], req[1], w.Code)
		}
	}
	if w := f.do("GET", "/api/dhcp", nil, c); w.Code != 200 {
		t.Errorf("reading stays possible: %d", w.Code)
	}
}

func TestPortScanAlertsOnNewlyOpenedPorts(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.seedDevice()
	f.withScope("192.168.1.0/24")
	var mu sync.Mutex
	var bodies []string
	hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		mu.Unlock()
	}))
	defer hook.Close()
	_ = f.srv.Alerts.SaveChannels(context.Background(), nil)
	if w := f.do("PUT", "/api/alerts", map[string]any{"rules": map[string]bool{"new_ports": true},
		"channels": []map[string]any{{"type": "webhook", "name": "h", "enabled": true, "url": hook.URL}}}, c); w.Code != 200 {
		t.Fatalf("alerts: %d %s", w.Code, w.Body)
	}
	open := map[string]bool{"192.168.1.10:22": true}
	f.srv.Dial = func(ctx context.Context, addr string) (net.Conn, error) {
		if open[addr] {
			a, _ := net.Pipe()
			return a, nil
		}
		return nil, errors.New("closed")
	}
	path := "/api/devices/b8:27:eb:00:00:01/portscan"
	f.do("POST", path, map[string]bool{"confirm": true}, c) // first scan: no previous result, no alert
	f.srv.lastScan = time.Time{}
	open["192.168.1.10:23"] = true
	f.do("POST", path, map[string]bool{"confirm": true}, c) // telnet opened
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(bodies)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 1 || !strings.Contains(bodies[0], "23") || !strings.Contains(bodies[0], "new open port") {
		t.Fatalf("expected one alert for port 23, got %v", bodies)
	}
}
