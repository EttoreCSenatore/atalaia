package fingerprint

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/identity"
)

// NetBIOSInfo is a device's NetBIOS node status.
type NetBIOSInfo struct {
	Name       string // computer name
	Workgroup  string
	FileServer bool // announces the file-server service (Windows, Samba, many NAS boxes)
}

// NetBIOS sends one NBSTAT query (UDP 137) straight to the device.
type NetBIOS struct {
	Port    int           // default 137
	Timeout time.Duration // default 700ms
}

// nbstatQuery is the wildcard node-status request: the name "*" padded with
// zeros and encoded half-ASCII, type NBSTAT, class IN.
func nbstatQuery(id uint16) []byte {
	q := make([]byte, 0, 50)
	q = binary.BigEndian.AppendUint16(q, id)
	q = append(q, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0) // flags, 1 question
	q = append(q, 0x20)
	q = append(q, "CKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"...)
	q = append(q, 0, 0, 0x21, 0, 1)
	return q
}

// Info asks the device. Found is false when it does not speak NetBIOS.
func (n NetBIOS) Info(ctx context.Context, ip netip.Addr) (NetBIOSInfo, bool) {
	port, timeout := n.Port, n.Timeout
	if port == 0 {
		port = 137
	}
	if timeout == 0 {
		timeout = 700 * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", netip.AddrPortFrom(ip, uint16(port)).String())
	if err != nil {
		return NetBIOSInfo{}, false
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	_ = conn.SetDeadline(deadline)
	const id = 0x4154
	if _, err := conn.Write(nbstatQuery(id)); err != nil {
		return NetBIOSInfo{}, false
	}
	buf := make([]byte, 1500)
	for {
		m, err := conn.Read(buf)
		if err != nil {
			return NetBIOSInfo{}, false
		}
		if info, ok := parseNBSTAT(buf[:m], id); ok {
			return info, true
		}
	}
}

// parseNBSTAT decodes a node-status answer: a 12-byte header, the echoed name
// (a 34-byte name or a 2-byte pointer), type, class, TTL, data length, then the
// name count and 18-byte entries (15-byte name, 1-byte suffix, 2 flag bytes).
func parseNBSTAT(p []byte, id uint16) (NetBIOSInfo, bool) {
	if len(p) < 12 || binary.BigEndian.Uint16(p[0:2]) != id || p[2]&0x80 == 0 || binary.BigEndian.Uint16(p[6:8]) == 0 {
		return NetBIOSInfo{}, false
	}
	i := 12
	if i < len(p) && p[i]&0xC0 == 0xC0 {
		i += 2
	} else {
		for i < len(p) && p[i] != 0 {
			i += int(p[i]) + 1
		}
		i++
	}
	i += 2 + 2 + 4 + 2 // type, class, ttl, rdlength
	if i >= len(p) {
		return NetBIOSInfo{}, false
	}
	rest := p[i:]
	count := int(rest[0])
	i++
	var info NetBIOSInfo
	found := false
	for k := 0; k < count && i+18 <= len(p); k++ {
		name := strings.TrimRight(string(p[i:i+15]), " \x00")
		suffix := p[i+15]
		flags := binary.BigEndian.Uint16(p[i+16 : i+18])
		group := flags&0x8000 != 0
		i += 18
		switch {
		case suffix == 0x00 && !group && info.Name == "":
			info.Name, found = identity.SanitizeLabel(name, 15), true
		case suffix == 0x00 && group && info.Workgroup == "":
			info.Workgroup = identity.SanitizeLabel(name, 15)
		case suffix == 0x20 && !group:
			info.FileServer = true
		}
	}
	return info, found
}
