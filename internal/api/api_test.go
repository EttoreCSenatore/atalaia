package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/alerts"
	"github.com/EttoreCSenatore/atalaia/internal/monitor"
	"github.com/EttoreCSenatore/atalaia/internal/notify"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

const (
	code     = "ABCD-EFGH-JKLM-NPQR"
	password = "a long enough password"
)

type fixture struct {
	t   *testing.T
	srv *Server
	h   http.Handler
	st  *store.Store
	mon *monitor.Monitor
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	mon := &monitor.Monitor{Store: st, Interval: time.Hour, Scan: func(context.Context, bool) ([]store.Observation, error) {
		return []store.Observation{{MAC: "b8:27:eb:00:00:01", IP: netip.MustParseAddr("192.168.1.10"), Vendor: "Raspberry Pi", Name: "pi.local"}}, nil
	}}
	sender := &notify.Sender{}
	srv := &Server{Store: st, Monitor: mon, Version: "test", Alerts: &alerts.Service{Store: st, Sender: sender}, Sender: sender}
	srv.SetSetupCode(code)
	return &fixture{t: t, srv: srv, h: srv.Handler(), st: st, mon: mon}
}

// do sends a request as a LAN client. mutate adds the CSRF header like the web app does.
func (f *fixture) do(method, path string, body any, cookie *http.Cookie, mods ...func(*http.Request)) *httptest.ResponseRecorder {
	f.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	r := httptest.NewRequest(method, path, rd)
	r.RemoteAddr = "192.168.1.50:5555"
	r.Host = "192.168.1.2:8080"
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		r.Header.Set(CSRFHeader, "1")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	for _, m := range mods {
		m(r)
	}
	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)
	return w
}

func session(w *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == cookieName {
			return c
		}
	}
	return nil
}

func (f *fixture) setup() *http.Cookie {
	f.t.Helper()
	w := f.do("POST", "/api/setup", map[string]string{"code": code, "password": password}, nil)
	if w.Code != 200 || session(w) == nil {
		f.t.Fatalf("setup: %d %s", w.Code, w.Body)
	}
	return session(w)
}

func TestFirstRunSetupFlow(t *testing.T) {
	f := newFixture(t)
	var st statusDTO
	_ = json.Unmarshal(f.do("GET", "/api/status", nil, nil).Body.Bytes(), &st)
	if !st.SetupRequired || st.Authenticated || st.Scan != nil {
		t.Fatalf("status before setup: %+v", st)
	}
	if w := f.do("GET", "/api/devices", nil, nil); w.Code != 401 {
		t.Fatalf("devices without login: %d", w.Code)
	}
	if w := f.do("POST", "/api/setup", map[string]string{"code": "WRONG-WRONG-WRONG-WRON", "password": password}, nil); w.Code != 403 {
		t.Fatalf("wrong code: %d", w.Code)
	}
	if w := f.do("POST", "/api/setup", map[string]string{"code": code, "password": "short"}, nil); w.Code != 400 {
		t.Fatalf("short password: %d", w.Code)
	}
	// the code is case-insensitive and tolerates spaces, since people type it
	w := f.do("POST", "/api/setup", map[string]string{"code": " " + strings.ToLower(code) + " ", "password": password}, nil)
	if w.Code != 200 {
		t.Fatalf("setup: %d %s", w.Code, w.Body)
	}
	c := session(w)
	if c == nil || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode || c.MaxAge <= 0 {
		t.Fatalf("cookie must be HttpOnly + SameSite=Strict: %+v", c)
	}
	if w := f.do("POST", "/api/setup", map[string]string{"code": code, "password": password}, nil); w.Code != 409 {
		t.Fatalf("second setup must be refused: %d", w.Code)
	}
	if w := f.do("GET", "/api/devices", nil, c); w.Code != 200 {
		t.Fatalf("devices after setup: %d", w.Code)
	}
}

func TestSetupDisabledWithoutCode(t *testing.T) {
	f := newFixture(t)
	f.srv.SetSetupCode("")
	if w := f.do("POST", "/api/setup", map[string]string{"code": "", "password": password}, nil); w.Code != 403 {
		t.Fatalf("empty code must never work: %d", w.Code)
	}
}

func TestLoginLogoutAndLockout(t *testing.T) {
	f := newFixture(t)
	f.setup()
	if w := f.do("POST", "/api/login", map[string]string{"password": "wrong password!!"}, nil); w.Code != 401 || session(w) != nil {
		t.Fatalf("wrong password: %d", w.Code)
	}
	w := f.do("POST", "/api/login", map[string]string{"password": password}, nil)
	c := session(w)
	if w.Code != 200 || c == nil {
		t.Fatalf("login: %d", w.Code)
	}
	if w := f.do("POST", "/api/logout", nil, c); w.Code != 200 {
		t.Fatalf("logout: %d", w.Code)
	}
	if w := f.do("GET", "/api/devices", nil, c); w.Code != 401 {
		t.Fatalf("the session must be dead after logout: %d", w.Code)
	}
	// brute force: 5 wrong passwords lock the client out, even for the right one
	for i := 0; i < 5; i++ {
		f.do("POST", "/api/login", map[string]string{"password": "wrong password!!"}, nil)
	}
	w = f.do("POST", "/api/login", map[string]string{"password": password}, nil)
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatalf("lockout: %d", w.Code)
	}
}

func TestPasswordChangeEndsOtherSessions(t *testing.T) {
	f := newFixture(t)
	a := f.setup()
	b := session(f.do("POST", "/api/login", map[string]string{"password": password}, nil))
	if w := f.do("POST", "/api/password", map[string]string{"current": "not the password", "new": "another long password"}, a); w.Code != 403 {
		t.Fatalf("wrong current: %d", w.Code)
	}
	if w := f.do("POST", "/api/password", map[string]string{"current": password, "new": "short"}, a); w.Code != 400 {
		t.Fatalf("weak new: %d", w.Code)
	}
	if w := f.do("POST", "/api/password", map[string]string{"current": password, "new": "another long password"}, a); w.Code != 200 {
		t.Fatalf("change: %d", w.Code)
	}
	if w := f.do("GET", "/api/devices", nil, a); w.Code != 200 {
		t.Fatal("the session that changed the password must survive")
	}
	if w := f.do("GET", "/api/devices", nil, b); w.Code != 401 {
		t.Fatal("other sessions must end")
	}
	if w := f.do("POST", "/api/login", map[string]string{"password": "another long password"}, nil); w.Code != 200 {
		t.Fatal("new password must work")
	}
}

func TestCSRFProtections(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	noHeader := func(r *http.Request) { r.Header.Del(CSRFHeader) }
	if w := f.do("POST", "/api/scan", nil, c, noHeader); w.Code != 403 {
		t.Fatalf("missing header: %d", w.Code)
	}
	evil := func(r *http.Request) { r.Header.Set("Origin", "http://evil.example") }
	if w := f.do("POST", "/api/scan", nil, c, evil); w.Code != 403 {
		t.Fatalf("foreign origin: %d", w.Code)
	}
	same := func(r *http.Request) { r.Header.Set("Origin", "http://192.168.1.2:8080") }
	if w := f.do("POST", "/api/scan", nil, c, same); w.Code != 202 {
		t.Fatalf("same origin: %d", w.Code)
	}
	if w := f.do("POST", "/api/login", map[string]string{"password": password}, nil, noHeader); w.Code != 403 {
		t.Fatalf("login must also need the header: %d", w.Code)
	}
}

func TestOnlyLocalClientsAndKnownHosts(t *testing.T) {
	f := newFixture(t)
	for remote, want := range map[string]int{
		"192.168.1.9:1": 200, "10.0.0.4:1": 200, "172.16.5.5:1": 200, "127.0.0.1:1": 200, "[::1]:1": 200, "[fe80::1]:1": 200, "[fd00::5]:1": 200,
		"8.8.8.8:1": 403, "[2001:db8::1]:1": 403, "203.0.113.7:1": 403, "garbage": 403,
	} {
		w := f.do("GET", "/api/status", nil, nil, func(r *http.Request) { r.RemoteAddr = remote })
		if w.Code != want {
			t.Errorf("client %s: %d, want %d", remote, w.Code, want)
		}
	}
	for host, want := range map[string]int{
		"192.168.1.2:8080": 200, "localhost:8080": 200, "[::1]:8080": 200, "evil.example.com": 403, "evil.example.com:8080": 403,
		"attacker.rebind.net": 403,
	} {
		w := f.do("GET", "/api/status", nil, nil, func(r *http.Request) { r.Host = host })
		if w.Code != want {
			t.Errorf("host %s: %d, want %d", host, w.Code, want)
		}
	}
	f.srv.AllowedHosts = []string{"atalaia.lan"}
	g := newFixture(t)
	g.srv.AllowedHosts = []string{"atalaia.lan"}
	if w := g.do("GET", "/api/status", nil, nil, func(r *http.Request) { r.Host = "Atalaia.LAN:8080" }); w.Code != 200 {
		t.Errorf("allowed host name: %d", w.Code)
	}
}

func TestSecurityHeadersAndErrors(t *testing.T) {
	f := newFixture(t)
	w := f.do("GET", "/api/status", nil, nil)
	for h, want := range map[string]string{
		"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Referrer-Policy": "no-referrer", "Cache-Control": "no-store",
	} {
		if w.Header().Get(h) != want {
			t.Errorf("%s = %q, want %q", h, w.Header().Get(h), want)
		}
	}
	if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") || !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Errorf("CSP = %q", csp)
	}
	if w := f.do("GET", "/api/nope", nil, nil); w.Code != 404 || !strings.Contains(w.Body.String(), "error") {
		t.Errorf("unknown API path: %d %s", w.Code, w.Body)
	}
	c := f.setup()
	// bad bodies
	r := f.do("POST", "/api/login", nil, nil, func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") })
	if r.Code != 415 {
		t.Errorf("wrong content type: %d", r.Code)
	}
	r = f.do("PUT", "/api/settings", map[string]any{"retention_days": 5, "extra": true}, c)
	if r.Code != 400 {
		t.Errorf("unknown fields must be refused: %d", r.Code)
	}
	r = f.do("POST", "/api/login", map[string]string{"password": strings.Repeat("x", 10000)}, nil)
	if r.Code != 400 {
		t.Errorf("oversized body must be refused: %d", r.Code)
	}
}

func TestDevicesEventsSettings(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	_, err := f.st.Record(context.Background(), time.Now(), []store.Observation{{MAC: "b8:27:eb:00:00:01", IP: netip.MustParseAddr("192.168.1.10"), Vendor: "Raspberry Pi", Name: "pi.local"}}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var devs []deviceDTO
	_ = json.Unmarshal(f.do("GET", "/api/devices", nil, c).Body.Bytes(), &devs)
	if len(devs) != 1 || devs[0].Name != "pi.local" || !devs[0].Online {
		t.Fatalf("devices: %+v", devs)
	}
	var evs []eventDTO
	_ = json.Unmarshal(f.do("GET", "/api/events?limit=5", nil, c).Body.Bytes(), &evs)
	if len(evs) != 1 || evs[0].Kind != "joined" {
		t.Fatalf("events: %+v", evs)
	}
	for _, bad := range []string{"0", "1001", "abc", "-3"} {
		if w := f.do("GET", "/api/events?limit="+bad, nil, c); w.Code != 400 {
			t.Errorf("limit=%s: %d", bad, w.Code)
		}
	}
	var set settingsDTO
	_ = json.Unmarshal(f.do("GET", "/api/settings", nil, c).Body.Bytes(), &set)
	if set.RetentionDays != monitor.DefaultRetentionDays {
		t.Fatalf("default retention: %+v", set)
	}
	if w := f.do("PUT", "/api/settings", map[string]int{"retention_days": 30}, c); w.Code != 200 {
		t.Fatalf("put settings: %d", w.Code)
	}
	if w := f.do("PUT", "/api/settings", map[string]int{"retention_days": 99999}, c); w.Code != 400 {
		t.Fatalf("absurd retention: %d", w.Code)
	}
	_ = json.Unmarshal(f.do("GET", "/api/settings", nil, c).Body.Bytes(), &set)
	if set.RetentionDays != 30 {
		t.Fatalf("retention: %+v", set)
	}
	var st statusDTO
	_ = json.Unmarshal(f.do("GET", "/api/status", nil, c).Body.Bytes(), &st)
	if !st.Authenticated || st.Scan == nil || st.SetupRequired {
		t.Fatalf("status after login: %+v", st)
	}
}

func TestStreamDeliversScanEvents(t *testing.T) {
	f := newFixture(t)
	ts := httptest.NewServer(f.h) // loopback client + 127.0.0.1 host are allowed
	defer ts.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	post := func(path, body string) *http.Response {
		req, _ := http.NewRequest("POST", ts.URL+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(CSRFHeader, "1")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}
	if r := post("/api/setup", `{"code":"`+code+`","password":"`+password+`"}`); r.StatusCode != 200 {
		t.Fatalf("setup: %d", r.StatusCode)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go f.mon.Run(ctx)

	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/stream", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("stream: %d %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	seen := map[string]bool{}
	sc := bufio.NewScanner(resp.Body)
	deadline := time.AfterFunc(5*time.Second, cancel)
	defer deadline.Stop()
	for sc.Scan() {
		if line := sc.Text(); strings.HasPrefix(line, "event: ") {
			seen[strings.TrimPrefix(line, "event: ")] = true
			if seen["hello"] && seen[monitor.ScanFinished] {
				return
			}
		}
	}
	t.Fatalf("did not receive hello + scan_finished, got %v", seen)
}

func TestStreamRequiresLogin(t *testing.T) {
	f := newFixture(t)
	if w := f.do("GET", "/api/stream", nil, nil); w.Code != 401 {
		t.Fatalf("stream without login: %d", w.Code)
	}
}
