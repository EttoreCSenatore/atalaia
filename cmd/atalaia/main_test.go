package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
	"github.com/EttoreCSenatore/atalaia/internal/tlsutil"
)

func newApp(t *testing.T, obs []store.Observation, now time.Time) (*app, *bytes.Buffer, *bytes.Buffer, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	a := &app{
		out: &out, err: &errOut,
		now:  func() time.Time { return now },
		scan: func(context.Context, bool) ([]store.Observation, error) { return obs, nil },
	}
	return a, &out, &errOut, filepath.Join(t.TempDir(), "atalaia.db")
}

func TestVersionAndUsage(t *testing.T) {
	a, out, errOut, _ := newApp(t, nil, time.Now())
	if code := a.run(context.Background(), []string{"version"}); code != 0 || !strings.HasPrefix(out.String(), "atalaia ") {
		t.Fatalf("version: code=%d out=%q", code, out.String())
	}
	if code := a.run(context.Background(), nil); code != 2 {
		t.Fatalf("no args: code=%d, want 2", code)
	}
	if code := a.run(context.Background(), []string{"bogus"}); code != 2 || !strings.Contains(errOut.String(), "unknown command") {
		t.Fatalf("unknown: code=%d", code)
	}
}

func TestScanDevicesHistory(t *testing.T) {
	t0 := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	obs := []store.Observation{
		{MAC: "b8:27:eb:00:00:01", IP: netip.MustParseAddr("192.168.1.10"), Vendor: "Raspberry Pi Trading Ltd", Name: "pi.local"},
		{MAC: "da:11:22:33:44:55", IP: netip.MustParseAddr("192.168.1.20"), Randomized: true},
	}
	a, out, _, db := newApp(t, obs, t0)

	if code := a.run(context.Background(), []string{"scan", "--db", db}); code != 0 {
		t.Fatalf("scan code=%d", code)
	}
	for _, want := range []string{"Found 2 devices: 2 new", "pi.local", "Raspberry Pi Trading Ltd", "(random MAC)", "online"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("scan output lacks %q:\n%s", want, out.String())
		}
	}

	out.Reset()
	if code := a.run(context.Background(), []string{"devices", "--json", "--db", db}); code != 0 {
		t.Fatalf("devices code=%d", code)
	}
	var devs []deviceJSON
	if err := json.Unmarshal(out.Bytes(), &devs); err != nil || len(devs) != 2 || devs[0].IP != "192.168.1.10" || !devs[1].Randomized {
		t.Fatalf("json: %v %+v", err, devs)
	}

	out.Reset()
	if code := a.run(context.Background(), []string{"history", "--db", db}); code != 0 || strings.Count(out.String(), "joined") != 2 {
		t.Fatalf("history code=%d out=%s", code, out.String())
	}
}

func TestEmptyDatabaseMessages(t *testing.T) {
	a, out, _, db := newApp(t, nil, time.Now())
	for _, c := range []string{"devices", "history"} {
		out.Reset()
		if code := a.run(context.Background(), []string{c, "--db", db}); code != 0 || !strings.Contains(out.String(), "atalaia scan") {
			t.Fatalf("%s: code=%d out=%q", c, code, out.String())
		}
	}
}

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *syncBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

func TestServeFirstRunAndShutdown(t *testing.T) {
	var out, errOut syncBuf
	a := &app{out: &out, err: &errOut, now: time.Now,
		scan: func(context.Context, bool) ([]store.Observation, error) { return nil, nil }}
	db := filepath.Join(t.TempDir(), "a.db")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() {
		done <- a.run(ctx, []string{"serve", "--listen", "127.0.0.1:0", "--interval", "0", "--db", db})
	}()

	var base string
	for i := 0; i < 100 && base == ""; i++ {
		if _, rest, ok := strings.Cut(out.String(), "listening on "); ok {
			base = strings.Fields(rest)[0]
		}
		time.Sleep(20 * time.Millisecond)
	}
	if base == "" {
		t.Fatalf("server did not start: %q %q", out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "Setup code") {
		t.Fatalf("first run must print the setup code: %q", out.String())
	}
	resp, err := http.Get(base + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var st struct {
		SetupRequired bool `json:"setup_required"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil || !st.SetupRequired {
		t.Fatalf("status: %v %+v", err, st)
	}
	if resp2, err := http.Get(base + "/"); err != nil || resp2.StatusCode != 200 {
		t.Fatalf("app: %v", err)
	} else {
		resp2.Body.Close()
	}
	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("exit code %d", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down")
	}
}

// startServe runs `serve` with args and returns the base URL, the console output and a stop function.
func startServe(t *testing.T, args ...string) (string, *syncBuf, func()) {
	t.Helper()
	var out, errOut syncBuf
	a := &app{out: &out, err: &errOut, now: time.Now,
		scan: func(context.Context, bool) ([]store.Observation, error) { return nil, nil }}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	go func() {
		done <- a.run(ctx, append([]string{"serve", "--listen", "127.0.0.1:0", "--interval", "0"}, args...))
	}()
	var base string
	for i := 0; i < 150 && base == ""; i++ {
		if _, rest, ok := strings.Cut(out.String(), "listening on "); ok {
			base = strings.Fields(rest)[0]
		}
		select {
		case code := <-done:
			t.Fatalf("serve exited early (%d): %q %q", code, out.String(), errOut.String())
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	if base == "" {
		cancel()
		t.Fatalf("server did not start: %q %q", out.String(), errOut.String())
	}
	return base, &out, func() { cancel(); <-done }
}

func insecureClient() *http.Client {
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
}

func TestServeSelfSignedHTTPS(t *testing.T) {
	dir := t.TempDir()
	base, out, stop := startServe(t, "--db", filepath.Join(dir, "a.db"), "--tls-self-signed")
	defer stop()
	if !strings.HasPrefix(base, "https://") || !strings.Contains(out.String(), "SHA-256") {
		t.Fatalf("base %q out %q", base, out.String())
	}
	resp, err := insecureClient().Get(base + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.TLS == nil || resp.TLS.Version < tls.VersionTLS12 {
		t.Fatalf("not TLS 1.2+: %+v", resp.TLS)
	}
	if resp.Header.Get("Strict-Transport-Security") != "" {
		t.Fatal("HSTS must not be sent with a self-signed certificate (it blocks the click-through)")
	}
	if st, err := os.Stat(filepath.Join(dir, "tls", "key.pem")); err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("key file: %v %v", st, err)
	}
	// plain HTTP to a TLS port does not work
	if r, err := http.Get(strings.Replace(base, "https://", "http://", 1) + "/api/status"); err == nil && r.StatusCode == 200 {
		t.Fatal("the TLS port answered plain HTTP")
	}
}

func TestServeWithOwnCertificateSendsHSTSAndSecureCookie(t *testing.T) {
	dir := t.TempDir()
	cert, key, err := tlsutil.SelfSigned(filepath.Join(dir, "mine"), []string{"localhost", "127.0.0.1"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	base, out, stop := startServe(t, "--db", filepath.Join(dir, "a.db"), "--tls-cert", cert, "--tls-key", key)
	defer stop()
	if strings.Contains(out.String(), "warning: serving the local network over plain HTTP") {
		t.Fatal("plain-HTTP warning shown on HTTPS")
	}
	resp, err := insecureClient().Get(base + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Strict-Transport-Security"), "max-age=") {
		t.Fatalf("no HSTS: %v", resp.Header)
	}
}

func TestServeTLSFlagValidation(t *testing.T) {
	for name, args := range map[string][]string{
		"cert only":  {"--tls-cert", "c.pem"},
		"key only":   {"--tls-key", "k.pem"},
		"both modes": {"--tls-self-signed", "--tls-cert", "c.pem", "--tls-key", "k.pem"},
		"missing":    {"--tls-cert", "/nonexistent/c.pem", "--tls-key", "/nonexistent/k.pem"},
	} {
		var out, errOut syncBuf
		a := &app{out: &out, err: &errOut, now: time.Now, scan: func(context.Context, bool) ([]store.Observation, error) { return nil, nil }}
		args = append([]string{"serve", "--listen", "127.0.0.1:0", "--db", filepath.Join(t.TempDir(), "a.db")}, args...)
		if code := a.run(context.Background(), args); code == 0 {
			t.Errorf("%s: expected a failure", name)
		}
	}
}

func TestServeExplainsAPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	var out, errOut syncBuf
	a := &app{out: &out, err: &errOut, now: time.Now, scan: func(context.Context, bool) ([]store.Observation, error) { return nil, nil }}
	code := a.run(context.Background(), []string{"serve", "--listen", ln.Addr().String(), "--db", filepath.Join(t.TempDir(), "a.db")})
	if code == 0 || !strings.Contains(errOut.String(), "already in use") || !strings.Contains(errOut.String(), "--listen") {
		t.Fatalf("code %d: %q", code, errOut.String())
	}
}

func TestServeRefusesAnOverWideAllowCIDR(t *testing.T) {
	var out, errOut syncBuf
	a := &app{out: &out, err: &errOut, now: time.Now, scan: func(context.Context, bool) ([]store.Observation, error) { return nil, nil }}
	if code := a.run(context.Background(), []string{"serve", "--listen", "127.0.0.1:0", "--allow-cidr", "0.0.0.0/0", "--db", filepath.Join(t.TempDir(), "a.db")}); code == 0 {
		t.Fatal("0.0.0.0/0 must be refused")
	}
}

func TestExportAndBackupCommands(t *testing.T) {
	var out, errOut syncBuf
	dir := t.TempDir()
	db := filepath.Join(dir, "a.db")
	a := &app{out: &out, err: &errOut, now: time.Now, scan: func(context.Context, bool) ([]store.Observation, error) {
		return []store.Observation{{MAC: "aa:00:00:00:00:01", IP: netip.MustParseAddr("192.168.1.10"), Name: "=1+1"}}, nil
	}}
	if code := a.run(context.Background(), []string{"scan", "--db", db}); code != 0 {
		t.Fatalf("scan: %q", errOut.String())
	}
	out = syncBuf{}
	if code := a.run(context.Background(), []string{"export", "--db", db}); code != 0 || !strings.Contains(out.String(), "'=1+1") || !strings.HasPrefix(out.String(), "mac,ip,") {
		t.Fatalf("csv: %d %q", code, out.String())
	}
	out = syncBuf{}
	if code := a.run(context.Background(), []string{"export", "--db", db, "--format", "json"}); code != 0 || !strings.Contains(out.String(), `"mac": "aa:00:00:00:00:01"`) {
		t.Fatalf("json: %q", out.String())
	}
	if code := a.run(context.Background(), []string{"export", "--db", db, "--format", "xml"}); code != 2 {
		t.Fatal("unknown format accepted")
	}
	bk := filepath.Join(dir, "copy.db")
	if code := a.run(context.Background(), []string{"backup", "--db", db, bk}); code != 0 {
		t.Fatalf("backup: %q", errOut.String())
	}
	if st, err := os.Stat(bk); err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("backup file: %v %v", st, err)
	}
	if code := a.run(context.Background(), []string{"backup", "--db", db, bk}); code == 0 {
		t.Fatal("overwrote an existing backup")
	}
	if code := a.run(context.Background(), []string{"backup", "--db", db}); code != 2 {
		t.Fatal("missing file name accepted")
	}
	// the backup time is recorded for /metrics and the heartbeat is pinged
	st, err := store.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	if last, err := st.LastBackup(context.Background()); err != nil || last.IsZero() {
		t.Fatalf("last backup not recorded: %v %v", last, err)
	}
	st.Close()
	var pings []string
	var mu sync.Mutex
	hc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		pings = append(pings, r.URL.Path)
		mu.Unlock()
	}))
	defer hc.Close()
	if code := a.run(context.Background(), []string{"backup", "--db", db, "--heartbeat", hc.URL + "/hc", filepath.Join(dir, "copy2.db")}); code != 0 {
		t.Fatalf("backup with heartbeat: %q", errOut.String())
	}
	if code := a.run(context.Background(), []string{"backup", "--db", db, "--heartbeat", hc.URL + "/hc", bk}); code == 0 {
		t.Fatal("overwrote an existing backup")
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(pings, " ") != "/hc /hc/fail" {
		t.Fatalf("pings = %v", pings)
	}
	if code := a.run(context.Background(), []string{"backup", "--db", db, "--heartbeat", "ftp://x", filepath.Join(dir, "copy3.db")}); code == 0 {
		t.Fatal("bad heartbeat URL accepted")
	}
}

func TestMetricsToken(t *testing.T) {
	t.Setenv("ATALAIA_METRICS_TOKEN", "")
	if tok, err := readMetricsToken(""); tok != "" || err != nil {
		t.Fatalf("off: %q %v", tok, err)
	}
	t.Setenv("ATALAIA_METRICS_TOKEN", "short")
	if _, err := readMetricsToken(""); err == nil {
		t.Fatal("short token accepted")
	}
	f := filepath.Join(t.TempDir(), "tok")
	if err := os.WriteFile(f, []byte(strings.Repeat("a", 40)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if tok, err := readMetricsToken(f); err != nil || tok != strings.Repeat("a", 40) {
		t.Fatalf("file: %q %v", tok, err)
	}
	// a systemd credential (LoadCredential=metrics-token:...)
	creds := t.TempDir()
	if err := os.WriteFile(filepath.Join(creds, "metrics-token"), []byte(strings.Repeat("c", 32)), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CREDENTIALS_DIRECTORY", creds)
	if tok, err := readMetricsToken(""); err != nil || tok != strings.Repeat("c", 32) {
		t.Fatalf("credential: %q %v", tok, err)
	}
}

func TestHealthCommand(t *testing.T) {
	var out, errOut syncBuf
	a := &app{out: &out, err: &errOut, now: time.Now}
	code := 200
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(code)
		if code == 200 {
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		} else {
			_, _ = w.Write([]byte(`{"status":"error","problems":["scans_stale"]}`))
		}
	}))
	defer srv.Close()
	if c := a.run(context.Background(), []string{"health", "--url", srv.URL}); c != 0 || strings.TrimSpace(out.String()) != "ok" {
		t.Fatalf("healthy: %d %q %q", c, out.String(), errOut.String())
	}
	code = 503
	out = syncBuf{}
	if c := a.run(context.Background(), []string{"health", "--url", srv.URL}); c != 1 || !strings.Contains(out.String(), "scans_stale") {
		t.Fatalf("unhealthy: %d %q", c, out.String())
	}
	if c := a.run(context.Background(), []string{"health", "--url", "http://127.0.0.1:1"}); c != 1 {
		t.Fatalf("unreachable: %d", c)
	}
	if c := a.run(context.Background(), []string{"health", "--url", "ftp://x"}); c != 2 {
		t.Fatalf("bad url: %d", c)
	}
	t.Setenv("ATALAIA_LISTEN", "0.0.0.0:8443")
	t.Setenv("ATALAIA_TLS_SELF_SIGNED", "1")
	if got := defaultHealthURL(); got != "https://127.0.0.1:8443" {
		t.Fatalf("default url = %s", got)
	}
}

func TestBackupDirAndVerify(t *testing.T) {
	var out, errOut syncBuf
	dir := t.TempDir()
	db := filepath.Join(dir, "a.db")
	a := &app{out: &out, err: &errOut, now: time.Now, scan: func(context.Context, bool) ([]store.Observation, error) {
		return []store.Observation{{MAC: "aa:00:00:00:00:01", IP: netip.MustParseAddr("192.168.1.10")}}, nil
	}}
	if code := a.run(context.Background(), []string{"scan", "--db", db}); code != 0 {
		t.Fatalf("scan: %q", errOut.String())
	}
	bk := filepath.Join(dir, "bk")
	if code := a.run(context.Background(), []string{"backup", "--db", db, "--dir", bk, "--keep", "2"}); code != 0 {
		t.Fatalf("backup --dir: %q", errOut.String())
	}
	if code := a.run(context.Background(), []string{"backup", "--db", db, "--dir", bk, filepath.Join(dir, "x.db")}); code != 2 {
		t.Fatal("both --dir and a file accepted")
	}
	out = syncBuf{}
	if code := a.run(context.Background(), []string{"verify-backup", "--dir", bk}); code != 0 || !strings.Contains(out.String(), "1 devices") {
		t.Fatalf("verify: %d %q %q", code, out.String(), errOut.String())
	}
	bad := filepath.Join(dir, "bad.db")
	if err := os.WriteFile(bad, []byte("not a db"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := a.run(context.Background(), []string{"verify-backup", bad}); code != 1 {
		t.Fatal("bad backup passed")
	}
	if code := a.run(context.Background(), []string{"verify-backup", "--dir", filepath.Join(dir, "none")}); code != 1 {
		t.Fatal("empty folder passed")
	}
}
