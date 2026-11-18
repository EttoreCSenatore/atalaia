// Package names finds the host name a device announces for its address:
// reverse DNS (what the router knows) and mDNS (what the device says about
// itself, e.g. "Marias-iPhone.local").
//
// Names come from the network and are untrusted: callers must sanitize them
// (identity.SanitizeLabel) before storing or printing.
package names

import (
	"context"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// Resolver returns a name for an address, or "" when it has none.
type Resolver interface {
	Name(ctx context.Context, ip netip.Addr) string
}

// Chain tries each resolver in order and returns the first non-empty name.
type Chain []Resolver

// Name implements Resolver.
func (c Chain) Name(ctx context.Context, ip netip.Addr) string {
	for _, r := range c {
		if n := r.Name(ctx, ip); n != "" {
			return n
		}
	}
	return ""
}

// ReverseDNS asks the system resolver for the PTR record of the address.
type ReverseDNS struct {
	Timeout time.Duration // default 1s
}

// Name implements Resolver.
func (r ReverseDNS) Name(ctx context.Context, ip netip.Addr) string {
	t := r.Timeout
	if t == 0 {
		t = time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()
	found, err := net.DefaultResolver.LookupAddr(ctx, ip.String())
	if err != nil || len(found) == 0 {
		return ""
	}
	return trimName(found[0])
}

// MDNS sends a reverse (PTR) query straight to the device's mDNS port. Devices
// answer such "legacy unicast" queries directly (RFC 6762 section 6.7), so no
// multicast is needed.
type MDNS struct {
	Port    int           // default 5353
	Timeout time.Duration // default 800ms
}

// Name implements Resolver.
func (m MDNS) Name(ctx context.Context, ip netip.Addr) string {
	port, timeout := m.Port, m.Timeout
	if port == 0 {
		port = 5353
	}
	if timeout == 0 {
		timeout = 800 * time.Millisecond
	}
	arpa, err := dns.ReverseAddr(ip.String())
	if err != nil {
		return ""
	}
	q := new(dns.Msg)
	q.SetQuestion(arpa, dns.TypePTR)
	q.RecursionDesired = false
	packet, err := q.Pack()
	if err != nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", netip.AddrPortFrom(ip, uint16(port)).String())
	if err != nil {
		return ""
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	_ = conn.SetDeadline(deadline)
	if _, err := conn.Write(packet); err != nil {
		return ""
	}
	buf := make([]byte, 1500)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return ""
		}
		resp := new(dns.Msg)
		if resp.Unpack(buf[:n]) != nil {
			continue
		}
		for _, rr := range resp.Answer {
			if ptr, ok := rr.(*dns.PTR); ok && strings.EqualFold(ptr.Hdr.Name, arpa) {
				return trimName(ptr.Ptr)
			}
		}
	}
}

func trimName(n string) string { return strings.TrimSuffix(strings.TrimSpace(n), ".") }

// ResolveAll names every address concurrently (at most workers at a time).
func ResolveAll(ctx context.Context, r Resolver, ips []netip.Addr, workers int) map[netip.Addr]string {
	if workers <= 0 {
		workers = 16
	}
	out := make(map[netip.Addr]string, len(ips))
	var mu sync.Mutex
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if n := r.Name(ctx, ip); n != "" {
				mu.Lock()
				out[ip] = n
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return out
}
