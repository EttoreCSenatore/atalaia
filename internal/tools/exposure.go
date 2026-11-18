package tools

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ---- risky services ---------------------------------------------------------

// Severity of a finding.
const (
	SevCritical = "critical"
	SevHigh     = "high"
	SevMedium   = "medium"
)

// Risk is an open port that should rarely be reachable on a home or office
// network. Key is the service name; the app explains it in the user's language.
type Risk struct {
	Port     int    `json:"port"`
	Service  string `json:"service"`
	Severity string `json:"severity"`
}

// riskyServices maps a service name (as in CommonPorts) to how serious it is
// to find it open. This is a rule table, not a vulnerability database.
var riskyServices = map[string]string{
	"docker-api": SevCritical, "telnet": SevHigh, "adb": SevHigh, "redis": SevHigh, "memcached": SevHigh,
	"mongodb": SevHigh, "elasticsearch": SevHigh, "couchdb": SevHigh, "kubernetes": SevHigh,
	"ftp": SevMedium, "smb": SevMedium, "rdp": SevMedium, "vnc": SevMedium, "mysql": SevMedium, "postgres": SevMedium,
	"mssql": SevMedium, "oracle": SevMedium, "rsync": SevMedium, "nfs": SevMedium, "msrpc": SevMedium,
	"netbios": SevMedium, "snmp": SevMedium, "rpcbind": SevMedium, "webmin": SevMedium, "socks": SevMedium, "pptp": SevMedium,
}

// Assess flags the open ports that are risky to expose.
func Assess(open []Port) []Risk {
	var out []Risk
	for _, p := range open {
		if sev, ok := riskyServices[p.Service]; ok {
			out = append(out, Risk{Port: p.Port, Service: p.Service, Severity: sev})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return sevRank(out[i].Severity) < sevRank(out[j].Severity) })
	return out
}

func sevRank(s string) int {
	switch s {
	case SevCritical:
		return 0
	case SevHigh:
		return 1
	}
	return 2
}

// ---- web panels -------------------------------------------------------------

// WebInfo is what a web port answers on "/".
type WebInfo struct {
	Port   int    `json:"port"`
	TLS    bool   `json:"tls"`
	Status int    `json:"status"`
	Server string `json:"server"`
	Title  string `json:"title"`
	Auth   bool   `json:"auth"` // asks for a login (401 / WWW-Authenticate)
	// Redirect is where a 3xx answer points (cleaned text; never followed).
	Redirect string `json:"redirect,omitempty"`
}

var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

var webPorts = map[int]bool{80: false, 81: false, 3000: false, 5000: false, 7000: false, 8000: false, 8008: false,
	8080: false, 8081: false, 8888: false, 9000: false, 9090: false, 10000: false, 32400: false, 631: false,
	443: true, 8443: true}

// IsWebPort reports whether Atalaia treats the port as HTTP (and whether it is TLS).
func IsWebPort(p int) (web, secure bool) {
	s, ok := webPorts[p]
	return ok, s
}

// TLSPorts are the open ports where a TLS handshake is tried.
var TLSPorts = map[int]bool{443: true, 8443: true, 465: true, 993: true, 995: true, 636: true, 8883: true}

// WebProbe sends one GET / to ip:port and reports status, Server header and
// page title. Redirects are not followed, at most 64 KiB are read, and the
// certificate is not verified (devices at home use self-signed ones; the TLS
// check reports that separately). Everything returned is plain, cleaned text.
func WebProbe(ctx context.Context, ip netip.Addr, port int, secure bool, dial Dialer) (WebInfo, bool) {
	info := WebInfo{Port: port, TLS: secure}
	scheme := "http"
	if secure {
		scheme = "https"
	}
	tr := &http.Transport{
		DisableKeepAlives:     true,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10}, // #nosec G402 -- probing the user's own device
		ResponseHeaderTimeout: 3 * time.Second,
	}
	if dial != nil {
		tr.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) { return dial(ctx, addr) }
	}
	cl := &http.Client{Transport: tr, Timeout: 6 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scheme+"://"+dialAddr(ip, port)+"/", nil)
	if err != nil {
		return info, false
	}
	req.Header.Set("User-Agent", "atalaia")
	res, err := cl.Do(req)
	if err != nil {
		return info, false
	}
	defer res.Body.Close()
	info.Status = res.StatusCode
	info.Server = clean(res.Header.Get("Server"), 100)
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		info.Redirect = clean(res.Header.Get("Location"), 200)
	}
	info.Auth = res.StatusCode == http.StatusUnauthorized || res.Header.Get("WWW-Authenticate") != ""
	body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if m := titleRe.FindSubmatch(body); m != nil {
		info.Title = clean(strings.Join(strings.Fields(string(m[1])), " "), 120)
	}
	return info, true
}

// ---- TLS --------------------------------------------------------------------

// TLSInfo is the result of a TLS handshake with a port.
type TLSInfo struct {
	Port        int    `json:"port"`
	Version     string `json:"version"`
	Subject     string `json:"subject"`
	Issuer      string `json:"issuer"`
	NotAfter    string `json:"not_after"`
	DaysLeft    int    `json:"days_left"`
	Expired     bool   `json:"expired"`
	SelfSigned  bool   `json:"self_signed"`
	LegacyTLS   bool   `json:"legacy_tls"`              // still accepts TLS 1.0/1.1
	WeakKeyBits int    `json:"weak_key_bits,omitempty"` // RSA key under 2048 bits
}

func versionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	}
	return "0x" + strconv.FormatUint(uint64(v), 16)
}

func handshake(ctx context.Context, ip netip.Addr, port int, dial Dialer, min, max uint16) (*tls.ConnectionState, bool) {
	if dial == nil {
		d := &net.Dialer{Timeout: 3 * time.Second}
		dial = func(ctx context.Context, addr string) (net.Conn, error) { return d.DialContext(ctx, "tcp4", addr) }
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	raw, err := dial(cctx, dialAddr(ip, port))
	if err != nil {
		return nil, false
	}
	defer raw.Close()
	c := tls.Client(raw, &tls.Config{InsecureSkipVerify: true, MinVersion: min, MaxVersion: max}) // #nosec G402 -- reads the certificate, trusts nothing
	if err := c.HandshakeContext(cctx); err != nil {
		return nil, false
	}
	st := c.ConnectionState()
	return &st, true
}

// TLSCheck performs a handshake with ip:port and describes the certificate and
// protocol. For protocols that start in clear text (465 is implicit TLS, so it
// works; STARTTLS ports are not covered) only implicit-TLS ports are checked.
func TLSCheck(ctx context.Context, ip netip.Addr, port int, dial Dialer, now time.Time) (TLSInfo, bool) {
	st, ok := handshake(ctx, ip, port, dial, tls.VersionTLS10, tls.VersionTLS13)
	if !ok {
		return TLSInfo{}, false
	}
	info := TLSInfo{Port: port, Version: versionName(st.Version)}
	if len(st.PeerCertificates) > 0 {
		describeCert(&info, st.PeerCertificates[0], now)
	}
	if _, legacy := handshake(ctx, ip, port, dial, tls.VersionTLS10, tls.VersionTLS11); legacy {
		info.LegacyTLS = true
	}
	return info, true
}

func describeCert(info *TLSInfo, c *x509.Certificate, now time.Time) {
	info.Subject = clean(c.Subject.String(), 200)
	info.Issuer = clean(c.Issuer.String(), 200)
	info.NotAfter = c.NotAfter.UTC().Format(time.RFC3339)
	info.Expired = now.After(c.NotAfter)
	info.DaysLeft = int(c.NotAfter.Sub(now).Hours() / 24)
	info.SelfSigned = c.Subject.String() == c.Issuer.String() && c.CheckSignature(c.SignatureAlgorithm, c.RawTBSCertificate, c.Signature) == nil
	if k, ok := c.PublicKey.(interface{ Size() int }); ok && c.PublicKeyAlgorithm == x509.RSA {
		if bits := k.Size() * 8; bits < 2048 {
			info.WeakKeyBits = bits
		}
	}
}
