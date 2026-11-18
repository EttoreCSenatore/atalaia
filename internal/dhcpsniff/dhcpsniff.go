// Package dhcpsniff learns device names from the DHCP requests that devices
// broadcast when they join or renew: most phones, even those with a random MAC,
// put their host name (option 12) and a vendor class (option 60) in them.
// It only listens; it never answers.
package dhcpsniff

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/identity"
)

// Client is what one device announced in a DHCP request.
type Client struct {
	MAC         string // chaddr, lowercase, colon separated
	Hostname    string // option 12
	VendorClass string // option 60, e.g. "android-dhcp-14", "MSFT 5.0"
}

var cookie = []byte{99, 130, 83, 99}

// Parse decodes a client DHCP message (DISCOVER, REQUEST or INFORM). ok is
// false for anything else, so replies from servers and malformed packets are ignored.
func Parse(p []byte) (Client, bool) {
	if len(p) < 241 || p[0] != 1 || p[1] != 1 || p[2] != 6 || string(p[236:240]) != string(cookie) {
		return Client{}, false
	}
	var c Client
	c.MAC = net.HardwareAddr(p[28:34]).String()
	isClient := false
	for i := 240; i < len(p); {
		code := p[i]
		if code == 255 {
			break
		}
		if code == 0 {
			i++
			continue
		}
		rest := p[i:]
		if len(rest) < 2 || 2+int(rest[1]) > len(rest) {
			return Client{}, false
		}
		n := int(rest[1])
		v := rest[2 : 2+n]
		switch code {
		case 53:
			isClient = n == 1 && (v[0] == 1 || v[0] == 3 || v[0] == 8)
		case 12:
			c.Hostname = identity.SanitizeLabel(string(v), 63)
		case 60:
			c.VendorClass = identity.SanitizeLabel(string(v), 63)
		}
		i += 2 + n
	}
	if !isClient || c.MAC == "00:00:00:00:00:00" {
		return Client{}, false
	}
	return c, true
}

// KindFromVendorClass maps a DHCP vendor class to a device kind ("" if unknown).
func KindFromVendorClass(vc string) string {
	v := strings.ToLower(vc)
	switch {
	case strings.HasPrefix(v, "android-dhcp"):
		return "phone"
	case strings.HasPrefix(v, "msft"):
		return "computer"
	case strings.HasPrefix(v, "udhcp"), strings.HasPrefix(v, "dhcpcd"):
		return "" // embedded and Linux clients alike: not enough to say
	}
	return ""
}

// Handler receives each client announcement.
type Handler func(ctx context.Context, c Client)

// Sniffer listens on the DHCP server port and calls the handler.
type Sniffer struct {
	Listen func() (net.PacketConn, error)
	Handle Handler

	mu     sync.Mutex
	err    string
	active bool
}

// Status reports whether the socket is open and the last error, if any.
func (s *Sniffer) Status() (active bool, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active, s.err
}

// Run keeps the socket open while enabled() is true, checking every `every`,
// until ctx ends. A failure to open it is remembered (see Status) and retried.
func (s *Sniffer) Run(ctx context.Context, enabled func() bool, every time.Duration, explain func(error) string) {
	var conn net.PacketConn
	closeConn := func() {
		if conn != nil {
			_ = conn.Close()
			conn = nil
		}
		s.mu.Lock()
		s.active = false
		s.mu.Unlock()
	}
	defer closeConn()
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		want := enabled()
		switch {
		case want && conn == nil:
			c, err := s.Listen()
			if err != nil {
				s.mu.Lock()
				s.err = explain(err)
				s.mu.Unlock()
			} else {
				conn = c
				s.mu.Lock()
				s.err, s.active = "", true
				s.mu.Unlock()
				go s.read(ctx, c) // ends when the socket is closed
			}
		case !want && conn != nil:
			closeConn()
			s.mu.Lock()
			s.err = ""
			s.mu.Unlock()
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func (s *Sniffer) read(ctx context.Context, c net.PacketConn) {
	buf := make([]byte, 1500)
	for ctx.Err() == nil {
		n, _, err := c.ReadFrom(buf)
		if err != nil {
			return // closed
		}
		if cl, ok := Parse(buf[:n]); ok && s.Handle != nil {
			s.Handle(ctx, cl)
		}
	}
}
