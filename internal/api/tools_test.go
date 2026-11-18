package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
	"github.com/EttoreCSenatore/atalaia/internal/store"
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

func TestExposureNeedsAPortScanConfirmationAndScope(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.seedDevice() // 192.168.1.10
	f.withScope("192.168.1.0/24")
	f.srv.Dial = func(context.Context, string) (net.Conn, error) { return nil, errors.New("closed") }
	path := "/api/devices/b8:27:eb:00:00:01/exposure"

	if w := f.do("POST", path, map[string]bool{"confirm": false}, c); w.Code != 400 {
		t.Fatalf("no confirmation: %d", w.Code)
	}
	if w := f.do("POST", path, map[string]bool{"confirm": true}, c); w.Code != 409 {
		t.Fatalf("no port scan yet: %d %s", w.Code, w.Body)
	}
	if w := f.do("GET", path, nil, c); w.Code != 404 {
		t.Fatalf("no report yet: %d", w.Code)
	}
	if err := f.st.SavePorts(context.Background(), "b8:27:eb:00:00:01", time.Now(), []store.OpenPort{{Port: 23, Service: "telnet"}, {Port: 22, Service: "ssh"}}); err != nil {
		t.Fatal(err)
	}
	w := f.do("POST", path, map[string]bool{"confirm": true}, c)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out exposureDTO
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out.Risks) != 1 || out.Risks[0].Service != "telnet" || out.Risks[0].Severity != "high" || out.Web == nil || out.TLS == nil {
		t.Fatalf("%+v", out)
	}
	if w := f.do("GET", path, nil, c); w.Code != 200 {
		t.Fatalf("stored report: %d", w.Code)
	}
	if w := f.do("POST", path, map[string]bool{"confirm": true}, c); w.Code != 429 {
		t.Fatalf("rate limit: %d", w.Code)
	}
	var a []auditDTO
	_ = json.Unmarshal(f.do("GET", "/api/audit", nil, c).Body.Bytes(), &a)
	if len(a) == 0 || a[0].Action != "exposure" {
		t.Fatalf("audit: %+v", a)
	}
	f.withScope("10.0.0.0/24")
	f.srv.lastScan = time.Time{}
	if w := f.do("POST", path, map[string]bool{"confirm": true}, c); w.Code != 403 {
		t.Fatalf("out of scope: %d", w.Code)
	}
}

func TestHTTPSMarksTheCookieSecureAndHSTSOnlyWhenAsked(t *testing.T) {
	for _, hsts := range []bool{false, true} {
		f := newFixture(t)
		f.srv.HSTS = hsts
		f.h = f.srv.Handler()
		w := f.do("POST", "/api/setup", map[string]string{"code": code, "password": password}, nil, func(r *http.Request) {
			r.TLS = &tls.ConnectionState{}
		})
		if w.Code != 200 {
			t.Fatalf("setup: %d %s", w.Code, w.Body)
		}
		c := session(w)
		if c == nil || !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
			t.Fatalf("cookie flags: %+v", c)
		}
		if got := w.Header().Get("Strict-Transport-Security") != ""; got != hsts {
			t.Fatalf("hsts=%v but header present=%v", hsts, got)
		}
	}
	// over plain HTTP: no HSTS, cookie not Secure (browsers would drop it otherwise)
	f := newFixture(t)
	f.srv.HSTS = true
	f.h = f.srv.Handler()
	w := f.do("POST", "/api/setup", map[string]string{"code": code, "password": password}, nil)
	if c := session(w); c == nil || c.Secure || w.Header().Get("Strict-Transport-Security") != "" {
		t.Fatalf("plain HTTP: %+v %v", c, w.Header())
	}
}

func TestSecureCookiesFlagForTLSProxies(t *testing.T) {
	f := newFixture(t)
	f.srv.SecureCookies = true
	f.h = f.srv.Handler()
	w := f.do("POST", "/api/setup", map[string]string{"code": code, "password": password}, nil)
	c := session(w)
	if c == nil || !c.Secure {
		t.Fatalf("cookie must be Secure behind a TLS proxy: %+v", c)
	}
	out := f.do("POST", "/api/logout", nil, c)
	if lc := session(out); lc == nil || !lc.Secure || lc.MaxAge >= 0 {
		t.Fatalf("logout cookie: %+v", lc)
	}
}

func TestObserveOnlyModeBlocksActiveToolsButNotDNSOrReading(t *testing.T) {
	f := newFixture(t)
	f.srv.NoActiveTools = true
	f.h = f.srv.Handler()
	c := f.setup()
	f.seedDevice()
	f.withScope("192.168.1.0/24")
	f.srv.Dial = func(context.Context, string) (net.Conn, error) {
		t.Fatal("dialed in observe-only mode")
		return nil, nil
	}
	mac := "b8:27:eb:00:00:01"
	for name, req := range map[string][2]any{
		"wake":     {"/api/devices/" + mac + "/wake", nil},
		"portscan": {"/api/devices/" + mac + "/portscan", map[string]bool{"confirm": true}},
		"exposure": {"/api/devices/" + mac + "/exposure", map[string]bool{"confirm": true}},
		"ping":     {"/api/tools/ping", map[string]any{"host": "192.168.1.1"}},
	} {
		if w := f.do("POST", req[0].(string), req[1], c); w.Code != 403 {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
	if w := f.do("GET", "/api/devices", nil, c); w.Code != 200 {
		t.Errorf("reading must still work: %d", w.Code)
	}
	raw := f.do("GET", "/api/status", nil, c).Body.String()
	if !strings.Contains(raw, `"active_tools":false`) {
		t.Fatalf("the web app reads active_tools from the status: %s", raw)
	}
	var st statusDTO
	_ = json.Unmarshal([]byte(raw), &st)
	if st.ActiveTools {
		t.Error("status must say active tools are off")
	}
	f2 := newFixture(t)
	var st2 statusDTO
	_ = json.Unmarshal(f2.do("GET", "/api/status", nil, nil).Body.Bytes(), &st2)
	if !st2.ActiveTools {
		t.Error("active tools are on by default")
	}
}

func TestSpeedTestNeedsConfirmationAndIsAudited(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	hits := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write(make([]byte, 64<<10))
	}))
	defer ts.Close()
	f.srv.SpeedURL = ts.URL
	if w := f.do("POST", "/api/tools/speedtest", map[string]bool{"confirm": false}, c); w.Code != 400 || hits != 0 {
		t.Fatalf("no confirmation: %d, hits %d", w.Code, hits)
	}
	w := f.do("POST", "/api/tools/speedtest", map[string]bool{"confirm": true}, c)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var r struct {
		Mbps  float64 `json:"mbps"`
		Bytes int64   `json:"bytes"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &r)
	if r.Bytes != 64<<10 || r.Mbps <= 0 {
		t.Fatalf("%+v", r)
	}
	var a []auditDTO
	_ = json.Unmarshal(f.do("GET", "/api/audit", nil, c).Body.Bytes(), &a)
	if len(a) != 1 || a[0].Action != "speedtest" {
		t.Fatalf("audit: %+v", a)
	}
	if w := f.do("POST", "/api/tools/speedtest", map[string]bool{"confirm": true}, nil); w.Code != 401 {
		t.Fatalf("anonymous: %d", w.Code)
	}
}

func TestTracerouteValidationAndObserveOnly(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	for _, bad := range []string{"", "224.0.0.1", "255.255.255.255"} {
		if w := f.do("POST", "/api/tools/traceroute", map[string]string{"host": bad}, c); w.Code != 400 {
			t.Errorf("%q: %d", bad, w.Code)
		}
	}
	if w := f.do("POST", "/api/tools/traceroute", map[string]string{"host": "127.0.0.1"}, c); w.Code != 200 {
		t.Fatalf("loopback: %d %s", w.Code, w.Body)
	}
	g := newFixture(t)
	g.srv.NoActiveTools = true
	g.h = g.srv.Handler()
	gc := g.setup()
	for _, p := range []string{"/api/tools/traceroute", "/api/tools/speedtest"} {
		if w := g.do("POST", p, map[string]any{"host": "127.0.0.1", "confirm": true}, gc); w.Code != 403 {
			t.Errorf("%s in observe-only mode: %d", p, w.Code)
		}
	}
}
