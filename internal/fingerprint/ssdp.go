package fingerprint

import (
	"context"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/identity"
	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

// SSDPInfo is what a UPnP device says about itself.
type SSDPInfo struct {
	Manufacturer string
	Model        string // model name
	Friendly     string // friendly name chosen by the owner or the maker
	DeviceType   string // e.g. "MediaRenderer", "InternetGatewayDevice"
	Server       string // SERVER header of the discovery answer
}

// SSDP discovers UPnP devices with one multicast M-SEARCH and reads each one's
// description. The description URL is only fetched when it points back at the
// very address that answered, inside the scan scope, over plain HTTP: a device
// cannot make Atalaia request an arbitrary address.
type SSDP struct {
	Dest  *net.UDPAddr // default 239.255.255.250:1900
	Wait  time.Duration
	Scope *scope.Scope
	// Fetch returns the body of a description URL (default: guarded HTTP GET).
	Fetch func(ctx context.Context, rawURL string) ([]byte, error)
}

// maxSSDPAnswers bounds how many devices one discovery round accepts.
const maxSSDPAnswers = 256

type ssdpAnswer struct {
	location, server string
}

// Discover returns what each responding address said about itself.
func (s SSDP) Discover(ctx context.Context) (map[netip.Addr]SSDPInfo, error) {
	dest := s.Dest
	if dest == nil {
		dest = &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1900}
	}
	wait := s.Wait
	if wait <= 0 {
		wait = 3 * time.Second
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	msg := "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 2\r\nST: ssdp:all\r\n\r\n"
	for i := 0; i < 2; i++ { // UDP is lossy: ask twice
		_, _ = conn.WriteToUDP([]byte(msg), dest)
		time.Sleep(100 * time.Millisecond)
	}
	_ = conn.SetReadDeadline(time.Now().Add(wait))
	answers := map[netip.Addr]ssdpAnswer{}
	buf := make([]byte, 4096)
	for ctx.Err() == nil {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		ip, ok := netip.AddrFromSlice(from.IP)
		if !ok || (s.Scope != nil && !s.Scope.Allows(ip.Unmap())) {
			continue
		}
		ip = ip.Unmap()
		if _, seen := answers[ip]; !seen && len(answers) >= maxSSDPAnswers {
			continue // a flood of (possibly spoofed) answers must not turn into a flood of requests
		}
		h := parseHeaders(string(buf[:n]))
		if loc := h["location"]; loc != "" {
			if _, seen := answers[ip]; !seen || answers[ip].location == "" {
				answers[ip] = ssdpAnswer{location: loc, server: h["server"]}
			}
		} else if _, seen := answers[ip]; !seen {
			answers[ip] = ssdpAnswer{server: h["server"]}
		}
	}

	fetch := s.Fetch
	if fetch == nil {
		fetch = guardedFetch
	}
	out := map[netip.Addr]SSDPInfo{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for ip, a := range answers {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			info := SSDPInfo{Server: identity.SanitizeLabel(a.server, 80)}
			if a.location != "" && locationOK(a.location, ip) {
				fctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				body, err := fetch(fctx, a.location)
				cancel()
				if err == nil {
					d := parseDescription(body)
					d.Server = info.Server
					info = d
				}
			}
			mu.Lock()
			out[ip] = info
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out, ctx.Err()
}

func parseHeaders(s string) map[string]string {
	h := map[string]string{}
	for i, line := range strings.Split(s, "\r\n") {
		if i == 0 {
			continue
		}
		if k, v, ok := strings.Cut(line, ":"); ok {
			h[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		}
	}
	return h
}

// locationOK allows only http://<the answering address>:<port>/path.
func locationOK(raw string, from netip.Addr) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.User != nil {
		return false
	}
	host, err := netip.ParseAddr(u.Hostname())
	return err == nil && host.Unmap() == from
}

func guardedFetch(ctx context.Context, raw string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "atalaia")
	cl := &http.Client{Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, io.EOF
	}
	return io.ReadAll(io.LimitReader(resp.Body, 128<<10))
}

type upnpRoot struct {
	Device struct {
		DeviceType   string `xml:"deviceType"`
		FriendlyName string `xml:"friendlyName"`
		Manufacturer string `xml:"manufacturer"`
		ModelName    string `xml:"modelName"`
		ModelDesc    string `xml:"modelDescription"`
	} `xml:"device"`
}

func parseDescription(b []byte) SSDPInfo {
	var r upnpRoot
	if xml.Unmarshal(b, &r) != nil { // encoding/xml never expands entities
		return SSDPInfo{}
	}
	d := r.Device
	model := d.ModelName
	if model == "" {
		model = d.ModelDesc
	}
	typ := d.DeviceType // urn:schemas-upnp-org:device:MediaRenderer:1
	if parts := strings.Split(typ, ":"); len(parts) >= 4 {
		typ = parts[len(parts)-2]
	}
	return SSDPInfo{
		Manufacturer: identity.SanitizeLabel(d.Manufacturer, 64),
		Model:        identity.SanitizeLabel(model, 64),
		Friendly:     identity.SanitizeLabel(d.FriendlyName, 64),
		DeviceType:   identity.SanitizeLabel(typ, 40),
	}
}
