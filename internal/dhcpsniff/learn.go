package dhcpsniff

import (
	"context"
	"sync"

	"github.com/EttoreCSenatore/atalaia/internal/fingerprint"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// Learner turns announcements into inventory updates. An announcement from a
// device the inventory does not know yet is kept (a few hundred at most) and
// applied after the next scan finds it.
type Learner struct {
	Store *store.Store

	mu      sync.Mutex
	pending map[string]Client
	learned int // how many devices got a name or a type from DHCP
}

const maxPending = 256

// Handle is a Handler.
func (l *Learner) Handle(ctx context.Context, c Client) {
	d, found, err := l.Store.DeviceByMAC(ctx, c.MAC)
	if err != nil {
		return
	}
	if !found {
		l.mu.Lock()
		if l.pending == nil {
			l.pending = map[string]Client{}
		}
		if len(l.pending) < maxPending {
			l.pending[c.MAC] = c
		}
		l.mu.Unlock()
		return
	}
	l.apply(ctx, d, c)
}

// Flush applies the announcements kept for devices that have since been found.
func (l *Learner) Flush(ctx context.Context) {
	l.mu.Lock()
	macs := make([]string, 0, len(l.pending))
	for m := range l.pending {
		macs = append(macs, m)
	}
	l.mu.Unlock()
	for _, m := range macs {
		d, found, err := l.Store.DeviceByMAC(ctx, m)
		if err != nil || !found {
			continue
		}
		l.mu.Lock()
		c := l.pending[m]
		delete(l.pending, m)
		l.mu.Unlock()
		l.apply(ctx, d, c)
	}
}

func (l *Learner) apply(ctx context.Context, d store.Device, c Client) {
	name := d.Name
	if name == "" {
		name = c.Hostname
	}
	kind := KindFromVendorClass(c.VendorClass)
	if g := fingerprint.Guess(fingerprint.GuessInput{Name: name, Vendor: d.Vendor, Model: d.Model, Services: d.Services, Randomized: d.Randomized}); kind == "" || (g != "" && g != fingerprint.KindMobile) {
		if g != "" && g != fingerprint.KindMobile {
			kind = g
		}
	}
	if changed, err := l.Store.UpdateIdentity(ctx, c.MAC, c.Hostname, kind); err == nil && changed {
		l.mu.Lock()
		l.learned++
		l.mu.Unlock()
	}
}

// Learned reports how many devices got a name or a type from DHCP.
func (l *Learner) Learned() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.learned
}
