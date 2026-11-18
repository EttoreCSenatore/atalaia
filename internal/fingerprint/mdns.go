// Package fingerprint learns what a device says it is: the mDNS services it
// advertises, its UPnP description (SSDP), its NetBIOS name. Everything it
// returns comes from the network, so it is untrusted: it is sanitized here,
// once, before anything else sees it.
package fingerprint

import (
	"context"
	"net"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/EttoreCSenatore/atalaia/internal/identity"
)

var serviceRe = regexp.MustCompile(`^_[a-z0-9][a-z0-9-]{0,30}\._(tcp|udp)$`)

// MDNSInfo is what a device's mDNS responder says about it.
type MDNSInfo struct {
	Services []string // service types, e.g. "_airplay._tcp"
	Model    string   // from the TXT records of the services it announces
}

// MDNS asks the device directly (legacy unicast, port 5353) which services it
// advertises, and for a few well-known ones reads the model from the TXT records.
type MDNS struct {
	Port    int           // default 5353
	Timeout time.Duration // per question, default 700ms
}

// interesting services whose TXT record holds a model, and the keys to look at.
var modelKeys = map[string][]string{
	"_googlecast._tcp":     {"md"},
	"_airplay._tcp":        {"model"},
	"_raop._tcp":           {"am"},
	"_device-info._tcp":    {"model"},
	"_ipp._tcp":            {"ty", "usb_MDL"},
	"_printer._tcp":        {"ty", "usb_MDL"},
	"_hap._tcp":            {"md"},
	"_companion-link._tcp": {"rpmd"},
}

// Info returns what the device at ip says about itself (empty if it says nothing).
func (m MDNS) Info(ctx context.Context, ip netip.Addr) MDNSInfo {
	var out MDNSInfo
	resp := m.ask(ctx, ip, "_services._dns-sd._udp.local.", dns.TypePTR)
	if resp == nil {
		return out
	}
	seen := map[string]bool{}
	for _, rr := range resp.Answer {
		ptr, ok := rr.(*dns.PTR)
		if !ok || !strings.EqualFold(ptr.Hdr.Name, "_services._dns-sd._udp.local.") {
			continue
		}
		t := strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(ptr.Ptr, "."), ".local"))
		if serviceRe.MatchString(t) && !seen[t] && len(out.Services) < 24 {
			seen[t] = true
			out.Services = append(out.Services, t)
		}
	}
	for _, t := range out.Services {
		keys, ok := modelKeys[t]
		if !ok || out.Model != "" {
			continue
		}
		out.Model = m.model(ctx, ip, t, keys)
	}
	return out
}

func (m MDNS) model(ctx context.Context, ip netip.Addr, service string, keys []string) string {
	r := m.ask(ctx, ip, service+".local.", dns.TypePTR)
	if r == nil {
		return ""
	}
	for _, rr := range r.Answer {
		ptr, ok := rr.(*dns.PTR)
		if !ok {
			continue
		}
		t := m.ask(ctx, ip, ptr.Ptr, dns.TypeTXT)
		if t == nil {
			continue
		}
		for _, a := range t.Answer {
			txt, ok := a.(*dns.TXT)
			if !ok {
				continue
			}
			for _, kv := range txt.Txt {
				k, v, found := strings.Cut(kv, "=")
				if !found {
					continue
				}
				for _, want := range keys {
					if strings.EqualFold(k, want) && strings.TrimSpace(v) != "" {
						return identity.SanitizeLabel(v, 64)
					}
				}
			}
		}
		break // one instance is enough
	}
	return ""
}

func (m MDNS) ask(ctx context.Context, ip netip.Addr, name string, qtype uint16) *dns.Msg {
	port, timeout := m.Port, m.Timeout
	if port == 0 {
		port = 5353
	}
	if timeout == 0 {
		timeout = 700 * time.Millisecond
	}
	q := new(dns.Msg)
	q.SetQuestion(dns.Fqdn(name), qtype)
	q.RecursionDesired = false
	packet, err := q.Pack()
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", netip.AddrPortFrom(ip, uint16(port)).String())
	if err != nil {
		return nil
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	_ = conn.SetDeadline(deadline)
	if _, err := conn.Write(packet); err != nil {
		return nil
	}
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil
		}
		resp := new(dns.Msg)
		if resp.Unpack(buf[:n]) == nil && resp.Id == q.Id && len(resp.Answer) > 0 {
			return resp
		}
	}
}
