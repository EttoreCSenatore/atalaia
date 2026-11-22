// Package monitor runs discovery scans in the background, on a schedule and on
// demand, and tells subscribers what happened.
package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// ScanFunc discovers the devices on the network.
type ScanFunc func(ctx context.Context, withNames bool) ([]store.Observation, error)

// Event kinds sent to subscribers.
const (
	ScanStarted  = "scan_started"
	ScanFinished = "scan_finished"
	ScanFailed   = "scan_failed"
)

// Event is something that happened to the monitor.
type Event struct {
	Kind    string    `json:"kind"`
	At      time.Time `json:"at"`
	Found   int       `json:"found,omitempty"`
	New     int       `json:"new,omitempty"`
	Online  int       `json:"came_online,omitempty"`
	Offline int       `json:"went_offline,omitempty"`
	Message string    `json:"message,omitempty"`
}

// Status is the monitor's current state.
type Status struct {
	Scanning  bool      `json:"scanning"`
	LastScan  time.Time `json:"last_scan"`
	LastError string    `json:"last_error"`
	Interval  string    `json:"interval"`
	// Duration of the last scan in seconds; Scans and Failures count scans since start.
	Duration float64 `json:"duration_seconds"`
	Scans    uint64  `json:"scans"`
	Failures uint64  `json:"failures"`
}

// RetentionKey is the setting that holds how many days of events are kept (0 = forever).
const RetentionKey = "retention_days"

// PruneKey holds after how many days offline an untouched device is forgotten (0 = never).
const PruneKey = "prune_days"

// DefaultRetentionDays is used until the user changes the setting.
const DefaultRetentionDays = 90

// Monitor owns the scan loop.
type Monitor struct {
	Store        *store.Store
	Scan         ScanFunc
	Interval     time.Duration // time between scans; <= 0 means only on demand
	OfflineAfter time.Duration
	Now          func() time.Time
	// OnResult, when set, is called after each recorded scan. baseline is true
	// when the inventory was empty before it (the first scan ever).
	OnResult func(ctx context.Context, res store.Result, baseline bool)
	// OnDone, when set, is called after every scan with its error (nil on
	// success), unless the monitor is stopping. Used for heartbeat pings.
	OnDone func(ctx context.Context, err error)

	trigger chan struct{}

	mu       sync.Mutex
	status   Status
	subs     map[chan Event]struct{}
	initOnce sync.Once
}

func (m *Monitor) init() {
	m.initOnce.Do(func() {
		m.trigger = make(chan struct{}, 1)
		m.subs = map[chan Event]struct{}{}
		if m.Now == nil {
			m.Now = time.Now
		}
		if m.OfflineAfter == 0 {
			m.OfflineAfter = 10 * time.Minute
		}
		m.status.Interval = m.Interval.String()
	})
}

// Trigger asks for a scan now. It never blocks; a request made while one is
// pending is merged into it.
func (m *Monitor) Trigger() {
	m.init()
	select {
	case m.trigger <- struct{}{}:
	default:
	}
}

// Status returns the current state.
func (m *Monitor) Status() Status {
	m.init()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

// Subscribe returns a channel of events and a function that ends the
// subscription. A subscriber that does not keep up loses events, never
// blocks the monitor.
func (m *Monitor) Subscribe() (<-chan Event, func()) {
	m.init()
	ch := make(chan Event, 16)
	m.mu.Lock()
	m.subs[ch] = struct{}{}
	m.mu.Unlock()
	return ch, func() {
		m.mu.Lock()
		if _, ok := m.subs[ch]; ok {
			delete(m.subs, ch)
			close(ch)
		}
		m.mu.Unlock()
	}
}

func (m *Monitor) publish(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for ch := range m.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Run scans immediately, then on every Interval or Trigger, until ctx ends.
func (m *Monitor) Run(ctx context.Context) {
	m.init()
	var tick <-chan time.Time
	if m.Interval > 0 {
		t := time.NewTicker(m.Interval)
		defer t.Stop()
		tick = t.C
	}
	m.once(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
		case <-m.trigger:
		}
		m.once(ctx)
	}
}

func (m *Monitor) once(ctx context.Context) {
	now := m.Now()
	m.mu.Lock()
	m.status.Scanning = true
	m.mu.Unlock()
	m.publish(Event{Kind: ScanStarted, At: now})

	obs, err := m.Scan(ctx, true)
	var res store.Result
	baseline := false
	if err == nil {
		if n, cerr := m.Store.CountDevices(ctx); cerr == nil {
			baseline = n == 0
		}
		res, err = m.Store.Record(ctx, now, obs, m.OfflineAfter)
	}
	if err == nil && m.OnResult != nil {
		m.OnResult(ctx, res, baseline)
	}
	if err == nil {
		m.purge(ctx, now)
	}

	m.mu.Lock()
	m.status.Scanning = false
	m.status.LastScan = now
	m.status.Duration = m.Now().Sub(now).Seconds()
	m.status.Scans++
	if err != nil {
		m.status.LastError = err.Error()
		m.status.Failures++
	} else {
		m.status.LastError = ""
	}
	m.mu.Unlock()
	if m.OnDone != nil && ctx.Err() == nil {
		m.OnDone(ctx, err)
	}

	if err != nil {
		if ctx.Err() == nil {
			m.publish(Event{Kind: ScanFailed, At: now, Message: err.Error()})
		}
		return
	}
	m.publish(Event{Kind: ScanFinished, At: now, Found: len(obs), New: len(res.New),
		Online: len(res.CameOnline), Offline: len(res.WentOffline)})
}

func (m *Monitor) purge(ctx context.Context, now time.Time) {
	days, err := m.Store.SettingInt(ctx, RetentionKey, DefaultRetentionDays)
	if err == nil && days > 0 {
		_, _ = m.Store.PurgeEvents(ctx, now.AddDate(0, 0, -days))
	}
	if keep, err := m.Store.SettingInt(ctx, PruneKey, 0); err == nil && keep > 0 {
		_, _ = m.Store.PruneDevices(ctx, now.AddDate(0, 0, -keep))
	}
	_ = m.Store.PurgeSessions(ctx, now)
}
