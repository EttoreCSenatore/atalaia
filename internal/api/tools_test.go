package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/netip"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

func (f *fixture) withScope(prefix string) {
	sc := scope.New([]netip.Prefix{netip.MustParsePrefix(prefix)})
	f.srv.Scope = func() (*scope.Scope, error) { return sc, nil }
}

func TestPortScanNeedsConfirmationScopeAndKeepsAudit(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.seedDevice() // 192.168.1.10
	f.withScope("192.168.1.0/24")
	var asked atomic.Int32
	f.srv.Dial = func(ctx context.Context, addr string) (net.Conn, error) {
		asked.Add(1)
		if addr == "192.168.1.10:22" {
			a, _ := net.Pipe()
			return a, nil
		}
		return nil, errors.New("closed")
	}
	path := "/api/devices/b8:27:eb:00:00:01/portscan"

	if w := f.do("POST", path, map[string]bool{"confirm": false}, c); w.Code != 400 || asked.Load() != 0 {
		t.Fatalf("without confirmation: %d, asked %d", w.Code, asked.Load())
	}
	if w := f.do("POST", path, map[string]bool{"confirm": true}, nil); w.Code != 401 {
		t.Fatalf("anonymous: %d", w.Code)
	}
	w := f.do("POST", path, map[string]bool{"confirm": true}, c)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out portsDTO
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out.Ports) != 1 || out.Ports[0].Port != 22 || out.Opened != nil {
		t.Fatalf("%+v", out)
	}
	// rate limit
	if w := f.do("POST", path, map[string]bool{"confirm": true}, c); w.Code != 429 {
		t.Fatalf("second scan right away: %d", w.Code)
	}
	// stored result and audit
	if w := f.do("GET", "/api/devices/b8:27:eb:00:00:01/ports", nil, c); w.Code != 200 {
		t.Fatalf("ports: %d", w.Code)
	}
	var a []auditDTO
	_ = json.Unmarshal(f.do("GET", "/api/audit", nil, c).Body.Bytes(), &a)
	if len(a) != 1 || a[0].Action != "portscan" || a[0].Target != "b8:27:eb:00:00:01" {
		t.Fatalf("audit: %+v", a)
	}

	// after the interval, a change is reported
	f.srv.lastScan = time.Now().Add(-time.Minute)
	f.srv.Dial = func(ctx context.Context, addr string) (net.Conn, error) {
		if addr == "192.168.1.10:80" {
			a, _ := net.Pipe()
			return a, nil
		}
		return nil, errors.New("closed")
	}
	w = f.do("POST", path, map[string]bool{"confirm": true}, c)
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out.Opened) != 1 || out.Opened[0] != 80 || len(out.Closed) != 1 || out.Closed[0] != 22 {
		t.Fatalf("diff: %+v", out)
	}
}

func TestActiveToolsRefuseDevicesOutsideScopeAndUnknownOnes(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.seedDevice()
	f.withScope("10.0.0.0/24") // the device is 192.168.1.10: out of scope
	f.srv.Dial = func(context.Context, string) (net.Conn, error) {
		t.Fatal("dialed an out-of-scope address")
		return nil, nil
	}
	if w := f.do("POST", "/api/devices/b8:27:eb:00:00:01/portscan", map[string]bool{"confirm": true}, c); w.Code != 403 {
		t.Fatalf("scan: %d", w.Code)
	}
	if w := f.do("POST", "/api/devices/b8:27:eb:00:00:01/wake", nil, c); w.Code != 403 {
		t.Fatalf("wake: %d", w.Code)
	}
	f.withScope("192.168.1.0/24")
	if w := f.do("POST", "/api/devices/aa:bb:cc:dd:ee:ff/portscan", map[string]bool{"confirm": true}, c); w.Code != 404 {
		t.Fatalf("unknown device: %d", w.Code)
	}
	if w := f.do("POST", "/api/devices/zz/wake", nil, c); w.Code != 400 {
		t.Fatalf("bad mac: %d", w.Code)
	}
}

func TestOneActiveToolAtATime(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.srv.active.Lock()
	defer f.srv.active.Unlock()
	if w := f.do("POST", "/api/tools/ping", map[string]any{"host": "192.168.1.1"}, c); w.Code != 429 {
		t.Fatalf("%d", w.Code)
	}
}

func TestPingAndDNSValidation(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	for name, body := range map[string]map[string]any{
		"multicast": {"host": "224.0.0.1"},
		"empty":     {"host": ""},
		"count":     {"host": "192.168.1.1", "count": 50},
	} {
		if w := f.do("POST", "/api/tools/ping", body, c); w.Code != 400 {
			t.Errorf("ping %s: %d %s", name, w.Code, w.Body)
		}
	}
	for name, body := range map[string]map[string]any{
		"type":  {"name": "example.com", "type": "ANY"},
		"empty": {"name": "", "type": "A"},
		"ptr":   {"name": "not-an-ip", "type": "PTR"},
	} {
		if w := f.do("POST", "/api/tools/dns", body, c); w.Code != 400 {
			t.Errorf("dns %s: %d %s", name, w.Code, w.Body)
		}
	}
	if w := f.do("POST", "/api/tools/dns", map[string]any{"name": "x.test", "type": "A"}, nil); w.Code != 401 {
		t.Errorf("anonymous dns: %d", w.Code)
	}
}
