// Package dhcpwatch runs the DHCP check: probe the network, remember the
// servers, alert about a new one.
package dhcpwatch

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/alerts"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpprobe"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// EnabledKey is the setting that turns the scheduled check on ("1").
const EnabledKey = "dhcp_enabled"

// ErrBusy means a check is already running.
var ErrBusy = errors.New("a DHCP check is already running")

// Report is the outcome of the last check.
type Report struct {
	At       time.Time         `json:"at"`
	Offers   []dhcpprobe.Offer `json:"offers"`
	Baseline bool              `json:"baseline"`
	New      []string          `json:"new"`
	Error    string            `json:"error,omitempty"`
}

// Watcher runs checks.
type Watcher struct {
	Store  *store.Store
	Alerts *alerts.Service
	// Probe sends the discover and returns the offers (see dhcpprobe.Probe).
	Probe func(ctx context.Context) ([]dhcpprobe.Offer, error)
	Now   func() time.Time

	mu   sync.Mutex
	last *Report
}

// Last returns the report of the last check, or nil.
func (w *Watcher) Last() *Report {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.last
}

// Enabled reports whether the scheduled check is on.
func (w *Watcher) Enabled(ctx context.Context) bool {
	n, err := w.Store.SettingInt(ctx, EnabledKey, 0)
	return err == nil && n != 0
}

// Check probes the network once. A probe that fails is recorded in the report
// (and returned), never as a baseline.
func (w *Watcher) Check(ctx context.Context) (Report, error) {
	if !w.mu.TryLock() {
		return Report{}, ErrBusy
	}
	defer w.mu.Unlock()
	now := time.Now()
	if w.Now != nil {
		now = w.Now()
	}
	rep := Report{At: now, Offers: []dhcpprobe.Offer{}, New: []string{}}
	offers, err := w.Probe(ctx)
	if err != nil {
		rep.Error = err.Error()
		w.last = &rep
		return rep, err
	}
	rep.Offers = offers
	obs := make([]store.DHCPObservation, 0, len(offers))
	for _, o := range offers {
		b, _ := json.Marshal(o)
		obs = append(obs, store.DHCPObservation{IP: o.Server.String(), Detail: string(b)})
	}
	res, err := w.Store.RecordDHCP(ctx, now, obs)
	if err != nil {
		rep.Error = "could not save the result"
		w.last = &rep
		return rep, err
	}
	rep.Baseline = res.Baseline
	for _, n := range res.New {
		rep.New = append(rep.New, n.IP)
	}
	w.last = &rep
	if len(res.New) > 0 && w.Alerts != nil {
		go w.Alerts.HandleDHCP(context.WithoutCancel(ctx), res.New)
	}
	return rep, nil
}
