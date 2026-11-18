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

// RoutePath is where Linux exposes the routing table.
const RoutePath = "/proc/net/route"

// DefaultGateway returns the IPv4 default gateway from a /proc/net/route file.
func DefaultGateway(path string) (netip.Addr, bool) {
	f, err := os.Open(path)
	if err != nil {
		return netip.Addr{}, false
	}
	defer f.Close()
	return ParseGateway(f)
}

// ParseGateway reads the default route (destination 0.0.0.0) of the
// /proc/net/route format. The address is stored as little-endian hex.
func ParseGateway(r io.Reader) (netip.Addr, bool) {
	sc := bufio.NewScanner(r)
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		f := strings.Fields(sc.Text())
		if len(f) < 8 || f[1] != "00000000" || f[7] != "00000000" {
			continue
		}
		flags, err := strconv.ParseUint(f[3], 16, 32)
		if err != nil || flags&0x3 != 0x3 { // up and gateway
			continue
		}
		v, err := strconv.ParseUint(f[2], 16, 32)
		if err != nil || v == 0 {
			continue
		}
		return netip.AddrFrom4([4]byte{byte(v & 0xff), byte(v >> 8 & 0xff), byte(v >> 16 & 0xff), byte(v >> 24 & 0xff)}), true
	}
	return netip.Addr{}, false
}

// SelfHosts returns the machine's own address on the main network: the
// neighbor table never lists the host itself, so without this the device
// running Atalaia would be missing from its own inventory. Only interfaces that
// are up and whose network contains the default gateway count (this leaves out
// Docker bridges and virtual interfaces).
func SelfHosts(gw netip.Addr, sc *scope.Scope) []Host {
	if !gw.IsValid() {
		return nil
	}
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	return selfHosts(gw, sc, ifs, func(i *net.Interface) ([]net.Addr, error) { return i.Addrs() })
}

func selfHosts(gw netip.Addr, sc *scope.Scope, ifs []net.Interface, addrs func(*net.Interface) ([]net.Addr, error)) []Host {
	var out []Host
	for i := range ifs {
		ifc := &ifs[i]
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 || len(ifc.HardwareAddr) != 6 || isZero(ifc.HardwareAddr) {
			continue
		}
		as, err := addrs(ifc)
		if err != nil {
			continue
		}
		for _, a := range as {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip, ok := netip.AddrFromSlice(ipn.IP)
			if !ok {
				continue
			}
			ip = ip.Unmap()
			ones, _ := ipn.Mask.Size()
			if !ip.Is4() || !netip.PrefixFrom(ip, ones).Contains(gw) || !sc.Allows(ip) {
				continue
			}
			out = append(out, Host{IP: ip, MAC: ifc.HardwareAddr, Iface: ifc.Name})
		}
	}
	return out
}
