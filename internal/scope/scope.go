// Package scope decides which addresses Atalaia may probe.
//
// It is the first safety guardrail of the project: active scans are limited to
// the private IPv4 networks the host itself is connected to. Nothing outside
// them (the internet, someone else's LAN) is ever a valid target.
package scope

import (
	"net"
	"net/netip"
	"sort"
)

// MaxPrefixBits is the shortest prefix (largest network) that is scanned.
// A /8 would mean 16 million addresses; larger networks are clamped to a /16
// that contains the host address.
const MaxPrefixBits = 16

// Scope is the set of networks that may be probed.
type Scope struct {
	prefixes []netip.Prefix
}

// New builds a Scope from candidate prefixes, keeping only private IPv4
// networks (RFC 1918), clamped to MaxPrefixBits, without duplicates.
func New(candidates []netip.Prefix) *Scope {
	seen := map[netip.Prefix]bool{}
	var kept []netip.Prefix
	for _, p := range candidates {
		addr := p.Addr().Unmap()
		if !addr.Is4() || !addr.IsPrivate() {
			continue
		}
		if p.Bits() < MaxPrefixBits {
			p = netip.PrefixFrom(addr, MaxPrefixBits)
		} else {
			p = netip.PrefixFrom(addr, p.Bits())
		}
		p = p.Masked()
		if !seen[p] {
			seen[p] = true
			kept = append(kept, p)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].String() < kept[j].String() })
	return &Scope{prefixes: kept}
}

// FromAddrs builds a Scope from interface addresses (as returned by
// net.Interface.Addrs).
func FromAddrs(addrs []net.Addr) *Scope {
	var candidates []netip.Prefix
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		addr, ok := netip.AddrFromSlice(ipnet.IP)
		if !ok {
			continue
		}
		ones, _ := ipnet.Mask.Size()
		candidates = append(candidates, netip.PrefixFrom(addr.Unmap(), ones))
	}
	return New(candidates)
}

// Local builds a Scope from the networks of the machine's interfaces that are
// up and not loopback.
func Local() (*Scope, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var all []net.Addr
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		all = append(all, addrs...)
	}
	return FromAddrs(all), nil
}

// Prefixes returns the networks in scope.
func (s *Scope) Prefixes() []netip.Prefix {
	return append([]netip.Prefix(nil), s.prefixes...)
}

// Allows reports whether addr may be probed: it must be inside one of the
// networks and must not be the network or broadcast address.
func (s *Scope) Allows(addr netip.Addr) bool {
	addr = addr.Unmap()
	for _, p := range s.prefixes {
		if !p.Contains(addr) {
			continue
		}
		if p.Bits() >= 31 { // /31 and /32 have no network/broadcast addresses
			return true
		}
		return addr != p.Masked().Addr() && addr != broadcast(p)
	}
	return false
}

func broadcast(p netip.Prefix) netip.Addr {
	a := p.Masked().Addr().As4()
	host := 32 - p.Bits()
	for i := 0; i < host; i++ {
		a[3-i/8] |= 1 << (i % 8)
	}
	return netip.AddrFrom4(a)
}
