// Package discovery finds the devices on the local network.
//
// The method needs no root and no raw sockets: Atalaia sends one tiny UDP
// datagram to every address of the local network, which makes the kernel
// resolve each address with ARP, and then reads the kernel's neighbor table.
// Every address is checked against the scan scope before it is probed and
// again before it is reported.
package discovery

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

// Host is a device seen on the network.
type Host struct {
	IP    netip.Addr
	MAC   net.HardwareAddr
	Iface string
}

// ARPTablePath is where Linux exposes the neighbor table.
const ARPTablePath = "/proc/net/arp"

// ErrUnsupported is returned on systems where the neighbor table cannot be read.
var ErrUnsupported = errors.New("device discovery is supported on Linux only for now")

// ParseARP parses the Linux /proc/net/arp format and returns the resolved
// entries (incomplete entries and all-zero MACs are skipped).
func ParseARP(r io.Reader) ([]Host, error) {
	var hosts []Host
	sc := bufio.NewScanner(r)
	first := true
	for sc.Scan() {
		if first { // header
			first = false
			continue
		}
		f := strings.Fields(sc.Text())
		if len(f) < 6 {
			continue
		}
		flags, err := strconv.ParseUint(strings.TrimPrefix(f[2], "0x"), 16, 32)
		if err != nil || flags&0x2 == 0 { // 0x2 = complete
			continue
		}
		ip, err := netip.ParseAddr(f[0])
		if err != nil {
			continue
		}
		mac, err := net.ParseMAC(f[3])
		if err != nil || isZero(mac) {
			continue
		}
		hosts = append(hosts, Host{IP: ip.Unmap(), MAC: mac, Iface: f[5]})
	}
	return hosts, sc.Err()
}

func isZero(mac net.HardwareAddr) bool {
	for _, b := range mac {
		if b != 0 {
			return false
		}
	}
	return true
}

// Scanner runs the sweep and reads the neighbor table.
type Scanner struct {
	// ARPPath is the neighbor table file (default ARPTablePath).
	ARPPath string
	// Nudge makes the kernel resolve ip (default: send one UDP datagram).
	Nudge func(ctx context.Context, ip netip.Addr)
	// Workers is the number of concurrent nudges (default 64).
	Workers int
	// Wait is how long to let ARP replies arrive after the sweep (default 2s).
	Wait time.Duration
}

// Scan probes the sweep targets of the scope and returns the hosts found,
// sorted by address. Hosts outside the scope are never returned.
func (s *Scanner) Scan(ctx context.Context, sc *scope.Scope) ([]Host, error) {
	path, nudge, workers, wait := s.ARPPath, s.Nudge, s.Workers, s.Wait
	if path == "" {
		path = ARPTablePath
	}
	if nudge == nil {
		nudge = udpNudge
	}
	if workers <= 0 {
		workers = 64
	}
	if wait == 0 {
		wait = 2 * time.Second
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrUnsupported, err)
	}

	targets := make(chan netip.Addr)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range targets {
				nudge(ctx, ip)
			}
		}()
	}
feed:
	for _, ip := range sc.SweepTargets() {
		select {
		case targets <- ip:
		case <-ctx.Done():
			break feed
		}
	}
	close(targets)
	wg.Wait()

	select {
	case <-time.After(wait):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all, err := ParseARP(f)
	if err != nil {
		return nil, err
	}
	var hosts []Host
	for _, h := range all {
		if sc.Allows(h.IP) {
			hosts = append(hosts, h)
		}
	}
	sortHosts(hosts)
	return hosts, nil
}

func sortHosts(h []Host) {
	for i := 1; i < len(h); i++ {
		for j := i; j > 0 && h[j].IP.Less(h[j-1].IP); j-- {
			h[j], h[j-1] = h[j-1], h[j]
		}
	}
}

// udpNudge sends one byte to the discard port; the datagram itself is
// irrelevant, the ARP resolution it triggers is what matters.
func udpNudge(ctx context.Context, ip netip.Addr) {
	var d net.Dialer
	d.Timeout = time.Second
	conn, err := d.DialContext(ctx, "udp4", netip.AddrPortFrom(ip, 9).String())
	if err != nil {
		return
	}
	_, _ = conn.Write([]byte{0})
	_ = conn.Close()
}
