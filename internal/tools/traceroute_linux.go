//go:build linux

package tools

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"golang.org/x/sys/unix"
)

// Hop is one step on the way to a host.
type Hop struct {
	TTL     int     `json:"ttl"`
	Addr    string  `json:"addr"`    // who answered ("" if nobody did)
	RTTms   float64 `json:"rtt_ms"`  // round trip to that hop
	Reached bool    `json:"reached"` // true for the destination itself
}

const (
	tracePortBase    = 33434
	seoeOriginICMP   = 2
	icmpTimeExceeded = 11
	icmpDestUnreach  = 3
)

// Traceroute finds the routers between this machine and dst. It needs no
// privileges: a UDP datagram with a small TTL is sent to a closed port and the
// ICMP "time exceeded" / "port unreachable" answers are read from the socket's
// error queue (IP_RECVERR). One probe per hop; a silent hop is reported as such.
func Traceroute(ctx context.Context, dst netip.Addr, maxHops int, perHop time.Duration) ([]Hop, error) {
	if !dst.Is4() {
		return nil, errors.New("traceroute supports IPv4 only")
	}
	if maxHops < 1 || maxHops > 30 {
		maxHops = 20
	}
	if perHop <= 0 {
		perHop = 2 * time.Second
	}
	var hops []Hop
	silent := 0
	for ttl := 1; ttl <= maxHops; ttl++ {
		if err := ctx.Err(); err != nil {
			return hops, err
		}
		h, err := probeHop(ctx, dst, ttl, perHop)
		if err != nil {
			return hops, err
		}
		hops = append(hops, h)
		if h.Reached {
			break
		}
		if h.Addr == "" {
			if silent++; silent >= 5 { // five silent hops in a row: the path is filtered, stop early
				break
			}
		} else {
			silent = 0
		}
	}
	return hops, nil
}

func probeHop(ctx context.Context, dst netip.Addr, ttl int, wait time.Duration) (Hop, error) {
	hop := Hop{TTL: ttl}
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return hop, errors.New("could not open a network socket")
	}
	defer unix.Close(fd)
	if err := unix.SetsockoptInt(fd, unix.SOL_IP, unix.IP_RECVERR, 1); err != nil {
		return hop, errors.New("this system does not allow reading ICMP errors")
	}
	if err := unix.SetsockoptInt(fd, unix.SOL_IP, unix.IP_TTL, ttl); err != nil {
		return hop, errors.New("could not set the TTL")
	}
	sa := &unix.SockaddrInet4{Port: tracePortBase + ttl, Addr: dst.As4()}
	start := time.Now()
	if err := unix.Sendto(fd, []byte("atalaia"), 0, sa); err != nil {
		return hop, nil // no route for this hop: shown as silent
	}
	deadline := start.Add(wait)
	for time.Now().Before(deadline) && ctx.Err() == nil {
		left := time.Until(deadline)
		if left > 100*time.Millisecond {
			left = 100 * time.Millisecond
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLERR}} // #nosec G115 -- a file descriptor fits int32
		n, err := unix.Poll(fds, int(left.Milliseconds())+1)
		if err != nil && err != unix.EINTR {
			return hop, nil
		}
		if n == 0 {
			continue
		}
		addr, typ, ok := readICMPError(fd)
		if !ok {
			continue
		}
		hop.Addr = addr.String()
		hop.RTTms = float64(time.Since(start).Microseconds()) / 1000
		// "port unreachable" from the destination ends the trace, and so does any
		// other "destination unreachable": there is nothing further to find.
		hop.Reached = typ == icmpDestUnreach
		return hop, nil
	}
	return hop, nil
}

// readICMPError reads one entry of the socket error queue and returns the
// address that sent the ICMP message and the ICMP type.
func readICMPError(fd int) (netip.Addr, uint8, bool) {
	p := make([]byte, 64)
	oob := make([]byte, 512)
	_, oobn, _, _, err := unix.Recvmsg(fd, p, oob, unix.MSG_ERRQUEUE|unix.MSG_DONTWAIT)
	if err != nil {
		return netip.Addr{}, 0, false
	}
	cmsgs, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return netip.Addr{}, 0, false
	}
	for _, c := range cmsgs {
		if c.Header.Level != unix.SOL_IP || c.Header.Type != unix.IP_RECVERR || len(c.Data) < 16+8 {
			continue
		}
		// struct sock_extended_err { u32 errno; u8 origin; u8 type; u8 code; u8 pad; u32 info; u32 data; }
		origin, typ := c.Data[4], c.Data[5]
		if origin != seoeOriginICMP {
			continue
		}
		// followed by the offender: struct sockaddr_in { u16 family; u16 port; u8 addr[4]; ... }
		off := c.Data[16:]
		a, ok := netip.AddrFromSlice(off[4:8])
		if !ok || a.IsUnspecified() {
			continue
		}
		return a, typ, true
	}
	return netip.Addr{}, 0, false
}
