package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
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
