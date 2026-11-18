// Package probe runs one real discovery pass: sweep the local networks, name
// the hosts, identify their vendors.
package probe

import (
	"context"
	"errors"
	"net/netip"
	"os"
	"sort"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/discovery"
	"github.com/EttoreCSenatore/atalaia/internal/fingerprint"
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
	gw, _ := discovery.DefaultGateway(discovery.RoutePath)
	hosts = appendSelf(hosts, discovery.SelfHosts(gw, sc))
	found := map[string]string{}
	if me, err := os.Hostname(); err == nil {
		for _, h := range discovery.SelfHosts(gw, sc) {
			found[h.IP.String()] = me
		}
	}
	if withNames {
		ips := make([]netip.Addr, 0, len(hosts))
		for _, h := range hosts {
			ips = append(ips, h.IP)
		}
		for ip, n := range names.ResolveAll(ctx, names.Chain{names.MDNS{}, names.ReverseDNS{}}, ips, 16) {
			if found[ip.String()] == "" {
				found[ip.String()] = n
			}
		}
	}
	reg := identity.Default()
	ips := make([]netip.Addr, 0, len(hosts))
	for _, h := range hosts {
		ips = append(ips, h.IP)
	}
	var learned map[netip.Addr]fingerprint.Result
	if withNames {
		learned = fingerprint.Gather(ctx, ips, sc, 6*time.Second)
	}
	return Collapse(hosts, gw, found, learned, reg), nil
}

// Collapse turns the hosts of a sweep into one observation per MAC. A MAC that
// answered for several addresses keeps all of them in IPs (the address shown is
// the gateway's if it has it, else the lowest) so the store can flag it.
func Collapse(hosts []discovery.Host, gateway netip.Addr, names map[string]string, learned map[netip.Addr]fingerprint.Result, reg *identity.DB) []store.Observation {
	index := map[string]int{}
	var obs []store.Observation
	for _, h := range hosts {
		mac := h.MAC.String()
		i, ok := index[mac]
		if !ok {
			index[mac] = len(obs)
			obs = append(obs, store.Observation{
				MAC:        mac,
				IP:         h.IP,
				Vendor:     identity.SanitizeLabel(reg.Vendor(h.MAC), 80),
				Randomized: identity.Randomized(h.MAC),
			})
			i = len(obs) - 1
		}
		o := &obs[i]
		o.IPs = append(o.IPs, h.IP)
		if gateway.IsValid() && h.IP == gateway {
			o.IP, o.Gateway = h.IP, true
		}
		if o.Name == "" {
			o.Name = identity.SanitizeLabel(names[h.IP.String()], 64)
		}
	}
	for i := range obs {
		o := &obs[i]
		var r fingerprint.Result
		for _, ip := range o.IPs { // the best-informed address of this MAC
			if x := learned[ip]; len(x.Services) > 0 || x.Model != "" || x.Manufacturer != "" || x.HasNetBIOS {
				r = x
				break
			}
		}
		if o.Name == "" && r.HasNetBIOS {
			o.Name = identity.SanitizeLabel(r.NetBIOS.Name, 64)
		}
		if o.Name == "" {
			o.Name = identity.SanitizeLabel(r.Friendly, 64)
		}
		if o.Vendor == "" {
			o.Vendor = identity.SanitizeLabel(r.Manufacturer, 80)
		}
		o.Model, o.Services = r.Model, r.Services
		o.Kind = fingerprint.Guess(fingerprint.GuessInput{
			Name: o.Name, Vendor: o.Vendor, Model: o.Model, Services: o.Services, DeviceType: r.DeviceType,
			NetBIOS: r.HasNetBIOS, Randomized: o.Randomized, Gateway: o.Gateway,
		})
	}
	return obs
}

// appendSelf adds the machine's own addresses unless the neighbor table already lists them.
func appendSelf(hosts, self []discovery.Host) []discovery.Host {
	seen := map[netip.Addr]bool{}
	for _, h := range hosts {
		seen[h.IP] = true
	}
	for _, h := range self {
		if !seen[h.IP] {
			hosts = append(hosts, h)
		}
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].IP.Less(hosts[j].IP) })
	return hosts
}
