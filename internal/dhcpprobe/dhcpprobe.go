// Package dhcpprobe looks for DHCP servers on the local network: it broadcasts
// one DHCPDISCOVER and listens for the OFFERs. A second DHCP server that the
// administrator did not set up (a "rogue" server) can hand out a wrong gateway
// or DNS and redirect everything. The probe never sends a REQUEST, so no lease
// is taken.
package dhcpprobe

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"net/netip"
	"os"
	"sort"
	"syscall"
	"time"
)

const (
	// ServerPort and ClientPort are the standard DHCP UDP ports.
	ServerPort = 67
	ClientPort = 68

	optMessageType = 53
	optServerID    = 54
	optRouter      = 3
	optDNS         = 6
	optLease       = 51
	optEnd         = 255
	typeDiscover   = 1
	typeOffer      = 2
)

var magicCookie = []byte{99, 130, 83, 99}

// Offer is what one DHCP server proposes.
type Offer struct {
	Server  netip.Addr   `json:"server"`  // DHCP server identifier (option 54)
	From    netip.Addr   `json:"from"`    // source address of the reply
	Offered netip.Addr   `json:"offered"` // address proposed to us
	Router  netip.Addr   `json:"router"`  // default gateway it would hand out
	DNS     []netip.Addr `json:"dns"`     // DNS servers it would hand out
	LeaseS  uint32       `json:"lease_s"` // lease time in seconds
}

// BuildDiscover builds a DHCPDISCOVER with a locally administered, random-looking
// client MAC (never the machine's own, so the server never ties a lease to it).
func BuildDiscover(xid uint32, mac net.HardwareAddr) []byte {
	p := make([]byte, 240, 300)
	p[0], p[1], p[2], p[3] = 1, 1, 6, 0 // BOOTREQUEST, ethernet, hlen, hops
	binary.BigEndian.PutUint32(p[4:8], xid)
	binary.BigEndian.PutUint16(p[10:12], 0x8000) // broadcast flag: the offer comes back as a broadcast
	copy(p[28:34], mac)
	copy(p[236:240], magicCookie)
	p = append(p, optMessageType, 1, typeDiscover)
	p = append(p, 55, 3, optRouter, optDNS, optLease) // parameter request list
	p = append(p, optEnd)
	return p
}

// RandomMAC returns a random locally administered unicast MAC address.
func RandomMAC() net.HardwareAddr {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	b[0] = b[0]&^0x01 | 0x02
	return b
}

// ParseOffer decodes a DHCPOFFER for the transaction xid. ok is false for
// anything else (other transactions, other message types, malformed packets).
func ParseOffer(p []byte, xid uint32, from netip.Addr) (Offer, bool) {
	if len(p) < 241 || p[0] != 2 || binary.BigEndian.Uint32(p[4:8]) != xid || string(p[236:240]) != string(magicCookie) {
		return Offer{}, false
	}
	o := Offer{From: from.Unmap()}
	if y, ok := netip.AddrFromSlice(p[16:20]); ok {
		o.Offered = y
	}
	isOffer := false
	for i := 240; i < len(p); {
		code := p[i]
		if code == optEnd {
			break
		}
		if code == 0 { // pad
			i++
			continue
		}
		rest := p[i:]
		if len(rest) < 2 || 2+int(rest[1]) > len(rest) {
			return Offer{}, false
		}
		n := int(rest[1])
		v := rest[2 : 2+n]
		switch code {
		case optMessageType:
			isOffer = n == 1 && v[0] == typeOffer
		case optServerID:
			if n == 4 {
				o.Server, _ = netip.AddrFromSlice(v)
			}
		case optRouter:
			if n >= 4 {
				o.Router, _ = netip.AddrFromSlice(v[:4])
			}
		case optDNS:
			for j := 0; j+4 <= n && len(o.DNS) < 4; j += 4 {
				if a, ok := netip.AddrFromSlice(v[j : j+4]); ok {
					o.DNS = append(o.DNS, a)
				}
			}
		case optLease:
			if n == 4 {
				o.LeaseS = binary.BigEndian.Uint32(v)
			}
		}
		i += 2 + n
	}
	if !isOffer {
		return Offer{}, false
	}
	if !o.Server.IsValid() {
		o.Server = o.From
	}
	return o, true
}

// Conn is the socket the probe uses (a *net.UDPConn in production).
type Conn interface {
	WriteTo(b []byte, addr net.Addr) (int, error)
	ReadFrom(b []byte) (int, net.Addr, error)
	SetReadDeadline(t time.Time) error
	Close() error
}

// Probe broadcasts one DHCPDISCOVER to dest and collects the OFFERs received
// within wait. Servers are deduplicated by identifier and sorted.
func Probe(ctx context.Context, conn Conn, dest *net.UDPAddr, wait time.Duration) ([]Offer, error) {
	if wait <= 0 {
		wait = 4 * time.Second
	}
	var xb [4]byte
	_, _ = rand.Read(xb[:])
	xid := binary.BigEndian.Uint32(xb[:])
	if _, err := conn.WriteTo(BuildDiscover(xid, RandomMAC()), dest); err != nil {
		return nil, errors.New("could not send the DHCP discover")
	}
	deadline := time.Now().Add(wait)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetReadDeadline(deadline)
	seen := map[netip.Addr]bool{}
	var out []Offer
	buf := make([]byte, 1500)
	for ctx.Err() == nil {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			break // deadline
		}
		var from netip.Addr
		if ua, ok := addr.(*net.UDPAddr); ok {
			from, _ = netip.AddrFromSlice(ua.IP)
		}
		if o, ok := ParseOffer(buf[:n], xid, from); ok && !seen[o.Server] {
			seen[o.Server] = true
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Server.Less(out[j].Server) })
	return out, ctx.Err()
}

// IsPermission reports whether err means the process may not bind the port.
func IsPermission(err error) bool {
	return errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) || errors.Is(err, os.ErrPermission)
}
