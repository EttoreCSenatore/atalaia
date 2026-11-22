package backup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/health"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

func TestScheduler(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var mu sync.Mutex
	var pings []string
	hc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		pings = append(pings, r.URL.Path)
		mu.Unlock()
	}))
	defer hc.Close()
	hb, _ := health.ParseHeartbeat(hc.URL + "/b")
	now := time.Date(2026, 11, 22, 3, 0, 0, 0, time.UTC)
	s := &Scheduler{Store: st, Dir: filepath.Join(dir, "backups"), Keep: 2, Interval: 24 * time.Hour, Heartbeat: hb, Now: func() time.Time { return now }}

	if (&Scheduler{}).Enabled() || (*Scheduler)(nil).Enabled() {
		t.Fatal("empty scheduler is enabled")
	}
	if got := (*Scheduler)(nil).State(ctx); got.Enabled {
		t.Fatal("nil state enabled")
	}
	// never backed up: late only after the grace period since start
	s.started = now
	if s.State(ctx).Late {
		t.Fatal("late right after start")
	}
	now = now.Add(3 * time.Hour)
	if !s.State(ctx).Late {
		t.Fatal("not late 3 h after start without a backup")
	}
	// first run: backup + restore test
	if err := s.RunIfDue(ctx); err != nil {
		t.Fatal(err)
	}
	state := s.State(ctx)
	if state.Last == nil || state.Late || state.LastRestoreTest == nil || !state.RestoreOK {
		t.Fatalf("after first run: %+v", state)
	}
	// not due yet: nothing happens
	now = now.Add(time.Hour)
	if err := s.RunIfDue(ctx); err != nil {
		t.Fatal(err)
	}
	if l, _ := store.Backups(s.Dir); len(l) != 1 {
		t.Fatalf("backups = %v", l)
	}
	// due after a day; restore test not repeated within the week
	for i := 0; i < 3; i++ {
		now = now.Add(24 * time.Hour)
		if err := s.RunIfDue(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if l, _ := store.Backups(s.Dir); len(l) != 2 {
		t.Fatalf("rotation: %v", l)
	}
	if at, _, _ := st.LastRestoreTest(ctx); !at.Equal(time.Date(2026, 11, 22, 6, 0, 0, 0, time.UTC)) {
		t.Fatalf("restore test repeated: %v", at)
	}
	// late after 26 h
	now = now.Add(27 * time.Hour)
	if !s.State(ctx).Late {
		t.Fatal("not late after 27 h")
	}
	mu.Lock()
	if strings.Join(pings, " ") != "/b /b /b /b" {
		t.Fatalf("pings = %v", pings)
	}
	mu.Unlock()
	// a failing backup pings /fail
	s.Dir = filepath.Join(dir, "a.db", "cannot")
	if err := s.RunNow(ctx); err == nil {
		t.Fatal("backup into a file path worked")
	}
	mu.Lock()
	if pings[len(pings)-1] != "/b/fail" {
		t.Fatalf("pings = %v", pings)
	}
	mu.Unlock()
}
