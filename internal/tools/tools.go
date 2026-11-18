// Package tools holds the network tools: ping, DNS lookup, Wake-on-LAN and a
// TCP port scan. Port scan and Wake-on-LAN only ever target a device address
// that the caller has already checked against the scan scope.
package tools

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ---- ping -------------------------------------------------------------------

// PingReply is one echo.
type PingReply struct {
	Seq   int     `json:"seq"`
	OK    bool    `json:"ok"`
	RTTms float64 `json:"rtt_ms"`
}

// PingResult is the outcome of a ping run.
type PingResult struct {
	Address string      `json:"address"`
	Sent    int         `json:"sent"`
	Replies []PingReply `json:"replies"`
	Lost    int         `json:"lost"`
}

// ErrPingUnavailable means unprivileged ICMP is not allowed on this host.
var ErrPingUnavailable = errors.New("ping needs unprivileged ICMP: set net.ipv4.ping_group_range to include the user running Atalaia")

// ResolveIPv4 turns an IPv4 address or a host name into one public-or-private
// unicast IPv4 address. Multicast, broadcast and unspecified addresses are refused.
func ResolveIPv4(ctx context.Context, host string) (netip.Addr, error) {
	host = strings.TrimSpace(host)
	if host == "" || len(host) > 253 {
		return netip.Addr{}, errors.New("invalid host")
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		var r net.Resolver
		ips, lerr := r.LookupNetIP(ctx, "ip4", host)
		if lerr != nil || len(ips) == 0 {
			return netip.Addr{}, errors.New("could not resolve the host name")
		}
		ip = ips[0]
	}
	ip = ip.Unmap()
	if !ip.Is4() || ip.IsMulticast() || ip.IsUnspecified() || ip == netip.AddrFrom4([4]byte{255, 255, 255, 255}) {
		return netip.Addr{}, errors.New("only unicast IPv4 addresses can be pinged")
	}
	return ip, nil
}

// Ping sends count echo requests (1 to 10) to ip.
func Ping(ctx context.Context, ip netip.Addr, count int, timeout time.Duration) (PingResult, error) {
	if count < 1 || count > 10 {
		count = 4
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	res := PingResult{Address: ip.String(), Sent: count}
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return res, ErrPingUnavailable
	}
	defer conn.Close()
	id := int(time.Now().UnixNano() & 0xffff)
	dst := &net.UDPAddr{IP: ip.AsSlice()}
	buf := make([]byte, 1500)
	for seq := 1; seq <= count; seq++ {
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		msg := icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: id, Seq: seq, Data: []byte("atalaia")}}
		b, _ := msg.Marshal(nil)
		start := time.Now()
		if _, err := conn.WriteTo(b, dst); err != nil {
			res.Replies = append(res.Replies, PingReply{Seq: seq})
			continue
		}
		reply := PingReply{Seq: seq}
		_ = conn.SetReadDeadline(start.Add(timeout))
		for {
			n, _, err := conn.ReadFrom(buf)
			if err != nil {
				break
			}
			m, err := icmp.ParseMessage(1, buf[:n])
			if err != nil || m.Type != ipv4.ICMPTypeEchoReply {
				continue
			}
			if e, ok := m.Body.(*icmp.Echo); ok && e.Seq == seq {
				reply.OK, reply.RTTms = true, float64(time.Since(start).Microseconds())/1000
				break
			}
		}
		if !reply.OK {
			res.Lost++
		}
		res.Replies = append(res.Replies, reply)
		if seq < count {
			select {
			case <-time.After(200 * time.Millisecond):
			case <-ctx.Done():
				return res, ctx.Err()
			}
		}
	}
	return res, nil
}

// ---- Wake-on-LAN ------------------------------------------------------------

// MagicPacket builds the Wake-on-LAN packet for mac: 6 x 0xFF then the MAC 16 times.
func MagicPacket(mac net.HardwareAddr) ([]byte, error) {
	if len(mac) != 6 {
		return nil, errors.New("Wake-on-LAN needs a 6-byte MAC address")
	}
	p := make([]byte, 0, 102)
	for i := 0; i < 6; i++ {
		p = append(p, 0xff)
	}
	for i := 0; i < 16; i++ {
		p = append(p, mac...)
	}
	return p, nil
}

// Wake broadcasts a magic packet for mac to the given broadcast address (UDP port 9).
func Wake(mac net.HardwareAddr, broadcast netip.Addr) error {
	p, err := MagicPacket(mac)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: broadcast.AsSlice(), Port: 9})
	if err != nil {
		return errors.New("could not open the network socket")
	}
	defer conn.Close()
	_, err = conn.Write(p)
	return err
}

// Broadcast returns the directed-broadcast address of prefix.
func Broadcast(p netip.Prefix) netip.Addr {
	a := p.Masked().Addr().As4()
	host := 32 - p.Bits()
	for i := 0; i < host; i++ {
		a[3-i/8] |= 1 << (i % 8)
	}
	return netip.AddrFrom4(a)
}

// ---- port scan --------------------------------------------------------------

// Port is an open TCP port.
type Port struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
}

// CommonPorts are the TCP ports checked, with the usual service name.
var CommonPorts = map[int]string{
	21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns", 80: "http", 81: "http-alt", 110: "pop3", 111: "rpcbind",
	123: "ntp", 135: "msrpc", 139: "netbios", 143: "imap", 161: "snmp", 389: "ldap", 443: "https", 445: "smb",
	465: "smtps", 515: "printer", 548: "afp", 554: "rtsp", 587: "submission", 631: "ipp", 636: "ldaps", 873: "rsync",
	993: "imaps", 995: "pop3s", 1080: "socks", 1433: "mssql", 1521: "oracle", 1723: "pptp", 1883: "mqtt",
	1900: "upnp", 2049: "nfs", 2375: "docker-api", 3000: "http-dev", 3306: "mysql", 3389: "rdp", 5000: "http-alt",
	5060: "sip", 5353: "mdns", 5432: "postgres", 5555: "adb", 5672: "amqp", 5900: "vnc", 5984: "couchdb",
	6379: "redis", 6443: "kubernetes", 7000: "http-alt", 8000: "http-alt", 8008: "http-alt", 8080: "http-proxy",
	8081: "http-alt", 8443: "https-alt", 8883: "mqtts", 8888: "http-alt", 9000: "http-alt", 9090: "http-alt",
	9100: "jetdirect", 9200: "elasticsearch", 10000: "webmin", 11211: "memcached", 27017: "mongodb", 32400: "plex",
	49152: "upnp",
}

// Dialer opens a TCP connection (replaced in tests).
type Dialer func(ctx context.Context, addr string) (net.Conn, error)

// PortScan checks CommonPorts on ip with at most `workers` connections at a time.
// Only completed TCP handshakes count as open; nothing is sent after connecting.
func PortScan(ctx context.Context, ip netip.Addr, dial Dialer, workers int, timeout time.Duration) []Port {
	if dial == nil {
		d := &net.Dialer{Timeout: timeout}
		dial = func(ctx context.Context, addr string) (net.Conn, error) { return d.DialContext(ctx, "tcp4", addr) }
	}
	if workers <= 0 {
		workers = 32
	}
	if timeout <= 0 {
		timeout = 700 * time.Millisecond
	}
	ports := make(chan int)
	var mu sync.Mutex
	var open []Port
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range ports {
				cctx, cancel := context.WithTimeout(ctx, timeout)
				c, err := dial(cctx, dialAddr(ip, p))
				cancel()
				if err != nil {
					continue
				}
				_ = c.Close()
				mu.Lock()
				open = append(open, Port{Port: p, Service: CommonPorts[p]})
				mu.Unlock()
			}
		}()
	}
	keys := make([]int, 0, len(CommonPorts))
	for p := range CommonPorts {
		keys = append(keys, p)
	}
	sort.Ints(keys)
feed:
	for _, p := range keys {
		select {
		case ports <- p:
		case <-ctx.Done():
			break feed
		}
	}
	close(ports)
	wg.Wait()
	sort.Slice(open, func(i, j int) bool { return open[i].Port < open[j].Port })
	return open
}

// ---- DNS --------------------------------------------------------------------

// DNSTypes are the record types DNSLookup answers.
var DNSTypes = []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "PTR"}

// DNSLookup resolves name with the system resolver. For PTR, name is an IP address.
// Answers are plain text, limited in number and length (they come from the network).
func DNSLookup(ctx context.Context, r *net.Resolver, name, typ string) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 253 {
		return nil, errors.New("invalid name")
	}
	if r == nil {
		r = net.DefaultResolver
	}
	var out []string
	var err error
	switch strings.ToUpper(typ) {
	case "A", "AAAA":
		var ips []netip.Addr
		network := "ip4"
		if strings.EqualFold(typ, "AAAA") {
			network = "ip6"
		}
		ips, err = r.LookupNetIP(ctx, network, name)
		for _, ip := range ips {
			out = append(out, ip.String())
		}
	case "CNAME":
		var c string
		c, err = r.LookupCNAME(ctx, name)
		if err == nil {
			out = []string{c}
		}
	case "MX":
		var mx []*net.MX
		mx, err = r.LookupMX(ctx, name)
		for _, m := range mx {
			out = append(out, fmt.Sprintf("%d %s", m.Pref, m.Host))
		}
	case "NS":
		var ns []*net.NS
		ns, err = r.LookupNS(ctx, name)
		for _, n := range ns {
			out = append(out, n.Host)
		}
	case "TXT":
		out, err = r.LookupTXT(ctx, name)
	case "PTR":
		ip, perr := netip.ParseAddr(name)
		if perr != nil {
			return nil, errors.New("PTR needs an IP address")
		}
		out, err = r.LookupAddr(ctx, ip.String())
	default:
		return nil, errors.New("unsupported record type")
	}
	if err != nil {
		var de *net.DNSError
		if errors.As(err, &de) && de.IsNotFound {
			return []string{}, nil
		}
		return nil, errors.New("the lookup failed")
	}
	if len(out) > 20 {
		out = out[:20]
	}
	for i, s := range out {
		out[i] = clean(s, 300)
	}
	return out, nil
}

func clean(s string, max int) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	if len(s) > max {
		s = s[:max]
	}
	return s
}

// dialAddr formats ip:port. A port outside 1-65535 cannot be dialed, so it is
// mapped to port 0, which every dial refuses.
func dialAddr(ip netip.Addr, port int) string {
	if port < 1 || port > 65535 {
		port = 0
	}
	return netip.AddrPortFrom(ip, uint16(port)).String() // #nosec G115 -- range checked above
}
