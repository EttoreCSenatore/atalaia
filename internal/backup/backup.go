// Package backup runs Atalaia's own backups while the server runs: a copy of
// the database every Interval into Dir (keeping the newest Keep), and a
// weekly restore test of the newest copy. It works the same with Docker, the
// packages and the plain binary.
package backup

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/health"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// RestoreEvery is how often the newest backup is test-restored.
const RestoreEvery = 7 * 24 * time.Hour

// Grace is how long past the interval a backup may be before it is "late"
// (a daily backup is late after 26 hours).
const Grace = 2 * time.Hour

// Scheduler makes the backups. A nil *Scheduler or Keep <= 0 means off.
type Scheduler struct {
	Store     *store.Store
	Dir       string
	Keep      int
	Interval  time.Duration
	Heartbeat *health.Heartbeat
	Now       func() time.Time
	Log       io.Writer

	mu      sync.Mutex // one backup at a time
	started time.Time
}

// State is what the app and the health endpoint show.
type State struct {
	Enabled         bool       `json:"enabled"`
	Dir             string     `json:"dir,omitempty"`
	Keep            int        `json:"keep,omitempty"`
	Interval        string     `json:"interval,omitempty"`
	Last            *time.Time `json:"last,omitempty"`
	Late            bool       `json:"late"`
	LastRestoreTest *time.Time `json:"last_restore_test,omitempty"`
	RestoreOK       bool       `json:"restore_ok"`
}

// Enabled reports whether scheduled backups are on.
func (s *Scheduler) Enabled() bool { return s != nil && s.Keep > 0 && s.Interval > 0 && s.Dir != "" }

func (s *Scheduler) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Run makes a backup when one is due, checking now and then every few
// minutes (so a machine that was asleep or off catches up), until ctx ends.
func (s *Scheduler) Run(ctx context.Context) {
	if !s.Enabled() {
		return
	}
	s.mu.Lock()
	s.started = s.now()
	s.mu.Unlock()
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for {
		if err := s.RunIfDue(ctx); err != nil && ctx.Err() == nil {
			s.logf("atalaia: scheduled backup failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// RunIfDue makes a backup if the last one is older than Interval.
func (s *Scheduler) RunIfDue(ctx context.Context) error {
	last, err := s.Store.LastBackup(ctx)
	if err != nil {
		return err
	}
	if !last.IsZero() && s.now().Sub(last) < s.Interval {
		return nil
	}
	return s.RunNow(ctx)
}

// RunNow makes a backup now and, if the last restore test is a week old or
// failed, test-restores it. The heartbeat (if any) hears about the outcome.
func (s *Scheduler) RunNow(ctx context.Context) error {
	if !s.Enabled() {
		return fmt.Errorf("scheduled backups are off")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.run(ctx)
	if perr := s.Heartbeat.Ping(context.WithoutCancel(ctx), err == nil); perr != nil {
		s.logf("atalaia: %v", perr)
	}
	return err
}

func (s *Scheduler) run(ctx context.Context) error {
	now := s.now()
	path, err := s.Store.BackupTo(ctx, s.Dir, s.Keep, now)
	if err != nil {
		return err
	}
	at, ok, err := s.Store.LastRestoreTest(ctx)
	if err != nil {
		return err
	}
	if ok && now.Sub(at) < RestoreEvery {
		return nil
	}
	_, verr := store.VerifyBackup(ctx, path, s.Dir)
	if err := s.Store.RecordRestoreTest(ctx, now, verr == nil); err != nil {
		return err
	}
	if verr != nil {
		return fmt.Errorf("restore test of %s: %w", filepath.Base(path), verr)
	}
	return nil
}

// State reports the backup state for the app and the health endpoint.
func (s *Scheduler) State(ctx context.Context) State {
	var st State
	if !s.Enabled() {
		return st
	}
	st.Enabled, st.Dir, st.Keep, st.Interval = true, s.Dir, s.Keep, s.Interval.String()
	now := s.now()
	if last, err := s.Store.LastBackup(ctx); err == nil && !last.IsZero() {
		st.Last = &last
		st.Late = now.Sub(last) > s.Interval+Grace
	} else {
		s.mu.Lock()
		started := s.started
		s.mu.Unlock()
		st.Late = !started.IsZero() && now.Sub(started) > Grace
	}
	if at, ok, err := s.Store.LastRestoreTest(ctx); err == nil && !at.IsZero() {
		st.LastRestoreTest, st.RestoreOK = &at, ok
	}
	return st
}

func (s *Scheduler) logf(format string, args ...any) {
	if s.Log != nil {
		fmt.Fprintf(s.Log, format+"\n", args...)
	}
}
