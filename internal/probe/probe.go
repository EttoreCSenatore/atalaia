// Package probe runs one real discovery pass: sweep the local networks, name
// the hosts, identify their vendors.
package probe

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/discovery"
	"github.com/EttoreCSenatore/atalaia/internal/identity"
	"github.com/EttoreCSenatore/atalaia/internal/names"
	"github.com/EttoreCSenatore/atalaia/internal/scope"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// Live sweeps the local networks, names the hosts it finds and identifies
// their vendors. Everything a device announces is sanitized here, once, so
// nothing downstream has to remember to do it.
func Live(ctx context.Context, withNames bool) ([]store.Observation, error) {
	sc, err := scope.Local()
	if err != nil {
		return nil, err
	}
	if len(sc.Prefixes()) == 0 {
		return nil, errors.New("no private IPv4 network found; nothing in scope")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	hosts, err := (&discovery.Scanner{}).Scan(ctx, sc)
	if err != nil {
		return nil, err
	}
	found := map[string]string{}
	if withNames {
		ips := make([]netip.Addr, 0, len(hosts))
		for _, h := range hosts {
			ips = append(ips, h.IP)
		}
		for ip, n := range names.ResolveAll(ctx, names.Chain{names.MDNS{}, names.ReverseDNS{}}, ips, 16) {
			found[ip.String()] = n
		}
	}
	reg := identity.Default()
	obs := make([]store.Observation, 0, len(hosts))
	for _, h := range hosts {
		obs = append(obs, store.Observation{
			MAC:        h.MAC.String(),
			IP:         h.IP,
			Vendor:     identity.SanitizeLabel(reg.Vendor(h.MAC), 80),
			Randomized: identity.Randomized(h.MAC),
			Name:       identity.SanitizeLabel(found[h.IP.String()], 64),
		})
	}
	return obs, nil
}
