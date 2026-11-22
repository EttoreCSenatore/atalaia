package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
)

func TestHealthWithoutLogin(t *testing.T) {
	f := newFixture(t)
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	f.srv.Now = func() time.Time { return now }
	f.mon.Now = func() time.Time { return now }

	read := func() (int, healthDTO) {
		w := f.do("GET", "/api/health", nil, nil)
		var h healthDTO
		if err := json.Unmarshal(w.Body.Bytes(), &h); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(w.Body.String(), "192.168") || strings.Contains(w.Body.String(), "b8:27") {
			t.Fatalf("health leaks network data: %s", w.Body)
		}
		return w.Code, h
	}
	// before any scan: fine
	if code, h := read(); code != 200 || h.Status != "ok" || h.Database != "ok" || h.Version != "test" || h.LastScan != nil {
		t.Fatalf("fresh: %d %+v", code, h)
	}
	// after a scan: fine, with the time
	go f.mon.Run(t.Context()) // one scan now, the next only after an hour
	waitFor(t, func() bool { return f.mon.Status().Scans > 0 })
	if code, h := read(); code != 200 || h.LastScan == nil {
		t.Fatalf("after scan: %d %+v", code, h)
	}
	// scans stopped for more than 3 intervals: 503
	now = now.Add(4 * time.Hour)
	if code, h := read(); code != 503 || h.Status != "error" || h.Problems[0] != "scans_stale" {
		t.Fatalf("stale: %d %+v", code, h)
	}
}

func TestHealthReportsFailedScanAndDatabase(t *testing.T) {
	f := newFixture(t)
	f.mon.Scan = func(context.Context, bool) ([]store.Observation, error) {
		return nil, errors.New("eth0 down 192.168.1.10")
	}
	go f.mon.Run(t.Context())
	waitFor(t, func() bool { return f.mon.Status().Failures > 0 })
	w := f.do("GET", "/api/health", nil, nil)
	if w.Code != 503 || !strings.Contains(w.Body.String(), "last_scan_failed") || strings.Contains(w.Body.String(), "eth0") {
		t.Fatalf("failed scan: %d %s", w.Code, w.Body)
	}
	f.st.Close()
	w = f.do("GET", "/api/health", nil, nil)
	if w.Code != 503 || !strings.Contains(w.Body.String(), `"database":"error"`) {
		t.Fatalf("closed db: %d %s", w.Code, w.Body)
	}
}

func TestMetricsNeedToken(t *testing.T) {
	f := newFixture(t)
	if w := f.do("GET", "/metrics", nil, nil); w.Code != http.StatusNotFound {
		t.Fatalf("disabled: %d", w.Code)
	}
	f.srv.MetricsToken = strings.Repeat("k", 40)
	bearer := func(tok string) func(*http.Request) {
		return func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+tok) }
	}
	if w := f.do("GET", "/metrics", nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", w.Code)
	}
	if w := f.do("GET", "/metrics", nil, nil, bearer("wrong")); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: %d", w.Code)
	}
	go f.mon.Run(t.Context())
	waitFor(t, func() bool { return f.mon.Status().Scans > 0 })
	if err := f.st.SetSetting(context.Background(), store.LastBackupKey, "1795000000"); err != nil {
		t.Fatal(err)
	}
	w := f.do("GET", "/metrics", nil, nil, bearer(f.srv.MetricsToken))
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain; version=0.0.4") {
		t.Fatalf("metrics: %d %s", w.Code, w.Header())
	}
	body := w.Body.String()
	for _, want := range []string{`atalaia_build_info{version="test"} 1`, "atalaia_devices 1\n", "atalaia_devices_online 1\n", "atalaia_scans_total 1\n",
		"atalaia_last_scan_success 1\n", "atalaia_last_backup_timestamp_seconds 1.795e+09\n", "atalaia_database_up 1\n"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
	if strings.Contains(body, "b8:27") || strings.Contains(body, "192.168") || strings.Contains(body, "pi.local") {
		t.Fatal("metrics leak network data")
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out")
}
