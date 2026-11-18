package monitor

import (
	"context"
	"errors"
	"net/netip"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func wait(t *testing.T, ch <-chan Event, kind string) Event {
	t.Helper()
	for {
		select {
		case e := <-ch:
			if e.Kind == kind {
				return e
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout waiting for %s", kind)
		}
	}
}

func TestRunScansOnStartAndOnTrigger(t *testing.T) {
	var calls atomic.Int32
	m := &Monitor{
		Store: newStore(t),
		Scan: func(context.Context, bool) ([]store.Observation, error) {
			calls.Add(1)
			return []store.Observation{{MAC: "a8:00:00:00:00:01", IP: netip.MustParseAddr("192.168.1.2")}}, nil
		},
	}
	events, cancelSub := m.Subscribe()
	defer cancelSub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)

	e := wait(t, events, ScanFinished)
	if e.Found != 1 || e.New != 1 {
		t.Fatalf("first scan: %+v", e)
	}
	m.Trigger()
	e = wait(t, events, ScanFinished)
	if e.New != 0 {
		t.Fatalf("second scan must find nothing new: %+v", e)
	}
	var st Status
	for i := 0; i < 100; i++ { // a merged second trigger may still be running
		if st = m.Status(); !st.Scanning {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if st.Scanning || st.LastScan.IsZero() || st.LastError != "" {
		t.Fatalf("status: %+v", st)
	}
}

func TestScanFailureIsReported(t *testing.T) {
	m := &Monitor{Store: newStore(t), Scan: func(context.Context, bool) ([]store.Observation, error) { return nil, errors.New("no network") }}
	events, cancelSub := m.Subscribe()
	defer cancelSub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)
	if e := wait(t, events, ScanFailed); e.Message != "no network" {
		t.Fatalf("%+v", e)
	}
	if m.Status().LastError != "no network" {
		t.Fatalf("status: %+v", m.Status())
	}
}

func TestSlowSubscriberNeverBlocksAndUnsubscribeCloses(t *testing.T) {
	m := &Monitor{Store: newStore(t), Scan: func(context.Context, bool) ([]store.Observation, error) { return nil, nil }}
	slow, cancelSlow := m.Subscribe()
	for i := 0; i < 100; i++ { // far more than the buffer, nobody reads
		m.publish(Event{Kind: ScanStarted})
	}
	cancelSlow()
	for range slow { // drains the buffered events; the loop only ends because the channel was closed
	}
	cancelSlow() // calling twice is harmless
}

func TestRetentionPurgesOldEvents(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	old := now.AddDate(0, 0, -100)
	_, _ = st.Record(ctx, old, []store.Observation{{MAC: "a8:00:00:00:00:01", IP: netip.MustParseAddr("192.168.1.2")}}, time.Hour)
	_ = st.SetSetting(ctx, RetentionKey, "30")
	m := &Monitor{Store: st, Now: func() time.Time { return now },
		Scan: func(context.Context, bool) ([]store.Observation, error) {
			return []store.Observation{{MAC: "a8:00:00:00:00:01", IP: netip.MustParseAddr("192.168.1.2")}}, nil
		}}
	m.init()
	m.once(ctx)
	evs, _ := st.Events(ctx, 10)
	for _, e := range evs {
		if e.At.Before(now.AddDate(0, 0, -31)) {
			t.Fatalf("old event survived: %+v", e)
		}
	}
}
