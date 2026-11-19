// Package api serves Atalaia's HTTP API and the embedded web app.
//
// Security model: the API is for the administrator of a network, so it is
// closed by default. Every request must come from a private or loopback
// address (never the internet), carry an acceptable Host header (against DNS
// rebinding) and, except for the first-run setup and login, a valid session
// cookie. State-changing requests also need a custom header, which a foreign
// web page cannot send without a CORS preflight that this server never grants.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/alerts"
	"github.com/EttoreCSenatore/atalaia/internal/auth"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpsniff"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpwatch"
	"github.com/EttoreCSenatore/atalaia/internal/monitor"
	"github.com/EttoreCSenatore/atalaia/internal/notify"
	"github.com/EttoreCSenatore/atalaia/internal/scope"
	"github.com/EttoreCSenatore/atalaia/internal/store"
	"github.com/EttoreCSenatore/atalaia/internal/tools"
)

const (
	cookieName = "atalaia_session"
	sessionTTL = 7 * 24 * time.Hour
	// CSRFHeader must be present on every state-changing request.
	CSRFHeader = "X-Atalaia"
	maxBody    = 4 << 10
	maxBigBody = 32 << 10 // alert channel list
	// maxStream bounds a live-update connection; the browser reconnects and
	// the session is checked again.
	maxStream = 30 * time.Minute
)

// Server is the HTTP API.
type Server struct {
	Store   *store.Store
	Monitor *monitor.Monitor
	Alerts  *alerts.Service
	DHCP    *dhcpwatch.Watcher
	Sniff   *dhcpsniff.Sniffer
	Learner *dhcpsniff.Learner
	// Scope returns the networks active tools may touch (default: the host's own private networks).
	Scope func() (*scope.Scope, error)
	// SpeedURL and SpeedClient replace the speed test file and HTTP client (tests).
	SpeedURL    string
	SpeedClient *http.Client
	// Dial opens TCP connections for the port scan (default: net dialer).
	Dial    tools.Dialer
	Sender  *notify.Sender
	Version string
	// AllowedNets are extra client networks treated as local (for example a VPN such as
	// Tailscale, 100.64.0.0/10), besides loopback, private and link-local addresses.
	AllowedNets []netip.Prefix
	// AllowedHosts are extra Host names accepted besides IP addresses,
	// localhost and the machine's own host name.
	AllowedHosts []string
	// NoActiveTools disables ping, Wake-on-LAN, port scan and the exposure check: the server then only observes.
	NoActiveTools bool
	// SecureCookies marks the session cookie Secure even over plain HTTP: for a TLS proxy in front of Atalaia.
	SecureCookies bool
	// HSTS adds Strict-Transport-Security on HTTPS responses (only for a certificate the user chose).
	HSTS bool
	// Static serves the web app for every path outside /api/ (may be nil).
	Static http.Handler
	Now    func() time.Time

	limiter auth.Limiter
	// global counts failures from every address together, so rotating
	// addresses on the LAN does not multiply the guesses an attacker gets.
	global auth.Limiter

	active   sync.Mutex // one active tool at a time
	lastScan time.Time  // last port scan, for the minimum interval

	mu         sync.Mutex
	setupCode  string
	hostnames  map[string]bool
	hostsReady sync.Once
}

// SetSetupCode sets the one-time first-run code ("" disables setup).
func (s *Server) SetSetupCode(code string) {
	s.mu.Lock()
	s.setupCode = code
	s.mu.Unlock()
}

func (s *Server) secureCookie(r *http.Request) bool { return r.TLS != nil || s.SecureCookies }

func (s *Server) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Handler returns the full handler with all protections applied.
func (s *Server) Handler() http.Handler {
	// all clients together: at most about 30 guesses a minute, and never a long
	// lockout (an attacker must not be able to lock the owner out for long)
	s.global.MaxFails, s.global.MaxWait = 30, time.Minute
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.status)
	mux.HandleFunc("POST /api/setup", s.mutating(s.setup))
	mux.HandleFunc("POST /api/login", s.mutating(s.login))
	mux.HandleFunc("POST /api/logout", s.mutating(s.protected(s.logout)))
	mux.HandleFunc("POST /api/password", s.mutating(s.protected(s.changePassword)))
	mux.HandleFunc("GET /api/devices", s.protected(s.devices))
	mux.HandleFunc("PUT /api/devices/{mac}", s.mutating(s.protected(s.updateDevice)))
	mux.HandleFunc("POST /api/devices/{mac}/wake", s.mutating(s.protected(s.activeOnly(s.wake))))
	mux.HandleFunc("POST /api/devices/{mac}/portscan", s.mutating(s.protected(s.activeOnly(s.portScan))))
	mux.HandleFunc("POST /api/devices/{mac}/exposure", s.mutating(s.protected(s.activeOnly(s.exposureScan))))
	mux.HandleFunc("GET /api/devices/{mac}/exposure", s.protected(s.exposure))
	mux.HandleFunc("GET /api/devices/{mac}/ports", s.protected(s.ports))
	mux.HandleFunc("POST /api/tools/ping", s.mutating(s.protected(s.activeOnly(s.ping))))
	mux.HandleFunc("POST /api/tools/traceroute", s.mutating(s.protected(s.activeOnly(s.traceroute))))
	mux.HandleFunc("POST /api/tools/speedtest", s.mutating(s.protected(s.activeOnly(s.speedTest))))
	mux.HandleFunc("POST /api/tools/dns", s.mutating(s.protected(s.dnsLookup)))
	mux.HandleFunc("GET /api/dhcp", s.protected(s.getDHCP))
	mux.HandleFunc("PUT /api/dhcp", s.mutating(s.protected(s.activeOnly(s.putDHCP))))
	mux.HandleFunc("PUT /api/dhcp/names", s.mutating(s.protected(s.putSniff)))
	mux.HandleFunc("POST /api/dhcp/check", s.mutating(s.protected(s.activeOnly(s.checkDHCP))))
	mux.HandleFunc("POST /api/dhcp/reset", s.mutating(s.protected(s.activeOnly(s.resetDHCP))))
	mux.HandleFunc("GET /api/audit", s.protected(s.audit))
	mux.HandleFunc("GET /api/alerts", s.protected(s.getAlerts))
	mux.HandleFunc("PUT /api/alerts", s.mutating(s.protected(s.putAlerts)))
	mux.HandleFunc("POST /api/alerts/test", s.mutating(s.protected(s.testAlert)))
	mux.HandleFunc("GET /api/events", s.protected(s.events))
	mux.HandleFunc("POST /api/scan", s.mutating(s.protected(s.scan)))
	mux.HandleFunc("GET /api/stream", s.protected(s.stream))
	mux.HandleFunc("GET /api/settings", s.protected(s.getSettings))
	mux.HandleFunc("PUT /api/settings", s.mutating(s.protected(s.putSettings)))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { writeError(w, http.StatusNotFound, "not found") })
	if s.Static != nil {
		mux.Handle("/", s.Static)
	}
	return s.securityHeaders(s.hostCheck(s.lanOnly(mux)))
}

// ---- middleware -----------------------------------------------------------

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; form-action 'self'; object-src 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		if s.HSTS && r.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=15552000")
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// lanOnly refuses requests that do not come from the local network: a port
// forwarded to the internet shows up here with a public source address.
func (s *Server) lanOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLocalClient(r.RemoteAddr) && !s.inAllowedNets(r.RemoteAddr) {
			writeError(w, http.StatusForbidden, "this server only answers to the local network")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) inAllowedNets(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	ip, err := netip.ParseAddr(strings.SplitN(host, "%", 2)[0])
	if err != nil {
		return false
	}
	ip = ip.Unmap()
	for _, p := range s.AllowedNets {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// ParseAllowedNets validates --allow-cidr values. A prefix that would let the
// whole internet in (anything shorter than /8 for IPv4, /32 for IPv6) is refused.
func ParseAllowedNets(list []string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, raw := range list {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		p, err := netip.ParsePrefix(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid network %q (use CIDR notation, e.g. 100.64.0.0/10)", raw)
		}
		if p.Addr().Is4() && p.Bits() < 8 || p.Addr().Is6() && p.Bits() < 32 {
			return nil, fmt.Errorf("network %q is too wide to be a local network", raw)
		}
		out = append(out, p.Masked())
	}
	return out, nil
}

func isLocalClient(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	ip := net.ParseIP(strings.TrimSuffix(strings.SplitN(host, "%", 2)[0], "."))
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

// hostCheck defends against DNS rebinding: a page on evil.example whose DNS
// name is switched to the server's address would otherwise be able to talk to
// it as if it were the same site.
func (s *Server) hostCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.hostAllowed(r.Host) {
			writeError(w, http.StatusForbidden, "unexpected Host header (use the server's IP address, localhost or an allowed name)")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) hostAllowed(hostport string) bool {
	s.hostsReady.Do(func() {
		s.hostnames = map[string]bool{"localhost": true}
		if hn, err := os.Hostname(); err == nil && hn != "" {
			hn = strings.ToLower(hn)
			s.hostnames[hn] = true
			s.hostnames[hn+".local"] = true
		}
		for _, h := range s.AllowedHosts {
			s.hostnames[strings.ToLower(strings.TrimSpace(h))] = true
		}
	})
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.ToLower(strings.Trim(host, "[]"))
	if net.ParseIP(host) != nil {
		return true
	}
	return s.hostnames[host]
}

// mutating requires the custom header and, when a browser sends an Origin, that
// it is this very server.
func (s *Server) mutating(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(CSRFHeader) == "" {
			writeError(w, http.StatusForbidden, "missing "+CSRFHeader+" header")
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || !strings.EqualFold(u.Host, r.Host) {
				writeError(w, http.StatusForbidden, "cross-origin request refused")
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) sessionHash(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" || len(c.Value) > 100 {
		return ""
	}
	return auth.HashToken(c.Value)
}

func (s *Server) protected(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authenticated(r) {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}
		next(w, r)
	}
}

func (s *Server) authenticated(r *http.Request) bool {
	h := s.sessionHash(r)
	if h == "" {
		return false
	}
	ok, err := s.Store.SessionValid(r.Context(), h, s.now())
	return err == nil && ok
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	return decodeMax(w, r, v, maxBody)
}

func decodeMax(w http.ResponseWriter, r *http.Request, v any, limit int64) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "send application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) startSession(w http.ResponseWriter, r *http.Request) error {
	token, hash, err := auth.NewToken()
	if err != nil {
		return err
	}
	if err := s.Store.CreateSession(r.Context(), hash, s.now(), sessionTTL); err != nil {
		return err
	}
	// #nosec G124 -- Secure follows the transport (see secureCookie): a Secure cookie over plain HTTP would be dropped by the browser
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: token, Path: "/", MaxAge: int(sessionTTL.Seconds()),
		HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: s.secureCookie(r),
	})
	return nil
}

func serverError(w http.ResponseWriter, err error) {
	// details stay out of the response; the server's own log has them
	fmt.Fprintln(os.Stderr, "atalaia: internal error:", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

// ---- handlers: account ------------------------------------------------------

type statusDTO struct {
	ActiveTools   bool            `json:"active_tools"`
	Version       string          `json:"version"`
	SetupRequired bool            `json:"setup_required"`
	Authenticated bool            `json:"authenticated"`
	Scan          *monitor.Status `json:"scan,omitempty"`
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	has, err := s.Store.HasUser(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	out := statusDTO{Version: s.Version, SetupRequired: !has, Authenticated: s.authenticated(r), ActiveTools: !s.NoActiveTools}
	if out.Authenticated {
		st := s.Monitor.Status()
		out.Scan = &st
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code     string `json:"code"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	ctx := r.Context()
	if has, err := s.Store.HasUser(ctx); err != nil {
		serverError(w, err)
		return
	} else if has {
		writeError(w, http.StatusConflict, "already set up")
		return
	}
	key := "setup|" + clientKey(r)
	if !s.attempt(w, key) {
		return
	}
	s.mu.Lock()
	code := s.setupCode
	s.mu.Unlock()
	if code == "" || !auth.EqualToken(strings.ToUpper(strings.TrimSpace(in.Code)), code) {
		writeError(w, http.StatusForbidden, "wrong setup code (it is printed in the server's console)")
		return
	}
	if err := auth.ValidatePassword(in.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.Store.CreateUser(ctx, hash, s.now()); err != nil {
		if errors.Is(err, store.ErrUserExists) {
			writeError(w, http.StatusConflict, "already set up")
			return
		}
		serverError(w, err)
		return
	}
	s.SetSetupCode("")
	s.limiter.Reset(key)
	s.global.Reset("all")
	if err := s.startSession(w, r); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// attempt counts a password or setup-code attempt for the client and for all
// clients together, before anything is checked. It answers 429 when either
// limit is reached.
func (s *Server) attempt(w http.ResponseWriter, key string) bool {
	now := s.now()
	if ok, wait := s.global.Allow("all", now); !ok {
		tooMany(w, wait)
		return false
	}
	if ok, wait := s.limiter.Attempt(key, now); !ok {
		tooMany(w, wait)
		return false
	}
	s.global.Fail("all", now)
	return true
}

func tooMany(w http.ResponseWriter, wait time.Duration) {
	w.Header().Set("Retry-After", fmt.Sprint(int(wait.Seconds())+1))
	writeError(w, http.StatusTooManyRequests, "too many attempts, try again later")
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	key := "login|" + clientKey(r)
	if !s.attempt(w, key) {
		return
	}
	hash, err := s.Store.PasswordHash(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	if hash == "" {
		writeError(w, http.StatusConflict, "setup required")
		return
	}
	if len(in.Password) > 1024 || !auth.VerifyPasswordLimited(r.Context(), hash, in.Password) {
		writeError(w, http.StatusUnauthorized, "wrong password")
		return
	}
	s.limiter.Reset(key)
	s.global.Reset("all")
	if err := s.startSession(w, r); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	_ = s.Store.DeleteSession(r.Context(), s.sessionHash(r))
	// #nosec G124 -- Secure follows the transport (see secureCookie): a Secure cookie over plain HTTP would be dropped by the browser
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: s.secureCookie(r)})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if !decode(w, r, &in) {
		return
	}
	key := "login|" + clientKey(r)
	if !s.attempt(w, key) {
		return
	}
	hash, err := s.Store.PasswordHash(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	if len(in.Current) > 1024 || !auth.VerifyPasswordLimited(r.Context(), hash, in.Current) {
		writeError(w, http.StatusForbidden, "wrong current password")
		return
	}
	if err := auth.ValidatePassword(in.New); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	newHash, err := auth.HashPassword(in.New)
	if err != nil {
		serverError(w, err)
		return
	}
	if err := s.Store.ChangePassword(r.Context(), newHash, s.sessionHash(r)); err != nil {
		serverError(w, err)
		return
	}
	s.limiter.Reset(key)
	s.global.Reset("all")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- handlers: data ---------------------------------------------------------

type deviceDTO struct {
	MAC        string   `json:"mac"`
	IP         string   `json:"ip"`
	Name       string   `json:"name"`
	Vendor     string   `json:"vendor"`
	Randomized bool     `json:"randomized_mac"`
	Online     bool     `json:"online"`
	FirstSeen  string   `json:"first_seen"`
	LastSeen   string   `json:"last_seen"`
	Label      string   `json:"label"`
	Tags       []string `json:"tags"`
	Known      bool     `json:"known"`
	Model      string   `json:"model"`
	Kind       string   `json:"kind"`
	Services   []string `json:"services"`
	This       bool     `json:"this_machine"`
}

func toDTO(d store.Device) deviceDTO {
	tags, services := d.Tags, d.Services
	if tags == nil {
		tags = []string{}
	}
	if services == nil {
		services = []string{}
	}
	return deviceDTO{d.MAC, d.IP, d.Name, d.Vendor, d.Randomized, d.Online,
		d.FirstSeen.Format(time.RFC3339), d.LastSeen.Format(time.RFC3339), d.Label, tags, d.Known, d.Model, d.Kind, services, false}
}

type eventDTO struct {
	At     string `json:"at"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
	MAC    string `json:"mac"`
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
}

func (s *Server) devices(w http.ResponseWriter, r *http.Request) {
	devs, err := s.Store.Devices(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	mine := localMACs()
	out := make([]deviceDTO, 0, len(devs))
	for _, d := range devs {
		dto := toDTO(d)
		dto.This = mine[d.MAC]
		out = append(out, dto)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n < 1 || n > 1000 {
			writeError(w, http.StatusBadRequest, "limit must be between 1 and 1000")
			return
		}
		limit = n
	}
	evs, err := s.Store.Events(r.Context(), limit)
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]eventDTO, 0, len(evs))
	for _, e := range evs {
		out = append(out, eventDTO{e.At.Format(time.RFC3339), e.Kind, e.Detail, e.MAC, e.Name, e.Vendor})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) scan(w http.ResponseWriter, r *http.Request) {
	s.Monitor.Trigger()
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "scheduled"})
}

// stream sends live updates (Server-Sent Events): scan started, finished or
// failed. The page reloads the device list when a scan finishes.
func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, cancel := s.Monitor.Subscribe()
	defer cancel()
	send := func(kind string, v any) bool {
		b, err := json.Marshal(v)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", kind, b); err != nil {
			return false
		}
		return rc.Flush() == nil
	}
	if !send("hello", s.Monitor.Status()) {
		return
	}
	ctx, stop := context.WithTimeout(r.Context(), maxStream)
	defer stop()
	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-events:
			if !ok || !send(e.Kind, e) {
				return
			}
		case <-keepalive.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil || rc.Flush() != nil {
				return
			}
		}
	}
}

type settingsDTO struct {
	RetentionDays int    `json:"retention_days"`
	ScanInterval  string `json:"scan_interval"`
	PortScanHours int    `json:"port_scan_hours"`
	ActiveTools   bool   `json:"active_tools"`
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	days, err := s.Store.SettingInt(r.Context(), monitor.RetentionKey, monitor.DefaultRetentionDays)
	if err != nil {
		serverError(w, err)
		return
	}
	hours, _ := s.Store.SettingInt(r.Context(), portScanHoursKey, 0)
	writeJSON(w, http.StatusOK, settingsDTO{RetentionDays: days, ScanInterval: s.Monitor.Status().Interval, PortScanHours: hours, ActiveTools: !s.NoActiveTools})
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RetentionDays *int `json:"retention_days"`
		PortScanHours *int `json:"port_scan_hours"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.RetentionDays == nil && in.PortScanHours == nil {
		writeError(w, http.StatusBadRequest, "nothing to change")
		return
	}
	if in.RetentionDays != nil && (*in.RetentionDays < 0 || *in.RetentionDays > 3650) {
		writeError(w, http.StatusBadRequest, "retention_days must be between 0 (keep forever) and 3650")
		return
	}
	if in.PortScanHours != nil {
		h := *in.PortScanHours
		if h < 0 || h > 720 {
			writeError(w, http.StatusBadRequest, "port_scan_hours must be between 0 (off) and 720")
			return
		}
		if h > 0 && s.NoActiveTools {
			writeError(w, http.StatusForbidden, "active tools are disabled on this server (observe-only mode)")
			return
		}
	}
	if in.RetentionDays != nil {
		if err := s.Store.SetSetting(r.Context(), monitor.RetentionKey, fmt.Sprint(*in.RetentionDays)); err != nil {
			serverError(w, err)
			return
		}
	}
	if in.PortScanHours != nil {
		if err := s.Store.SetSetting(r.Context(), portScanHoursKey, fmt.Sprint(*in.PortScanHours)); err != nil {
			serverError(w, err)
			return
		}
		s.audited(r, "portscan-schedule", "", fmt.Sprintf("every %d h", *in.PortScanHours))
	}
	s.getSettings(w, r)
}

func (s *Server) updateDevice(w http.ResponseWriter, r *http.Request) {
	hw, err := net.ParseMAC(r.PathValue("mac"))
	if err != nil || len(hw) != 6 {
		writeError(w, http.StatusBadRequest, "invalid MAC address")
		return
	}
	var in struct {
		Label string   `json:"label"`
		Tags  []string `json:"tags"`
		Known bool     `json:"known"`
	}
	if !decode(w, r, &in) {
		return
	}
	label, err := store.CleanLabel(in.Label)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tags, err := store.CleanTags(in.Tags)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	d, found, err := s.Store.UpdateDevice(r.Context(), hw.String(), label, tags, in.Known)
	if err != nil {
		serverError(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "unknown device")
		return
	}
	writeJSON(w, http.StatusOK, toDTO(d))
}

// ---- handlers: alerts -------------------------------------------------------

// channelDTO is a channel as shown to the browser: secrets are never sent
// back, only whether one is set.
type channelDTO struct {
	notify.Channel
	HasToken    bool `json:"has_token"`
	HasPassword bool `json:"has_password"`
}

type alertsDTO struct {
	Rules    alerts.Rules `json:"rules"`
	Channels []channelDTO `json:"channels"`
}

func redact(c notify.Channel) channelDTO {
	d := channelDTO{Channel: c, HasToken: c.Token != "", HasPassword: c.Password != ""}
	d.Channel.Token, d.Channel.Password = "", ""
	return d
}

func (s *Server) alertsView(w http.ResponseWriter, r *http.Request) {
	rules, err := s.Alerts.Rules(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	chans, err := s.Alerts.Channels(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	out := alertsDTO{Rules: rules, Channels: make([]channelDTO, 0, len(chans))}
	for _, c := range chans {
		out.Channels = append(out.Channels, redact(c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getAlerts(w http.ResponseWriter, r *http.Request) { s.alertsView(w, r) }

func (s *Server) putAlerts(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Rules    alerts.Rules `json:"rules"`
		Channels []channelDTO `json:"channels"` // the shape GET returns, so a view can be sent back
	}
	if !decodeMax(w, r, &in, maxBigBody) {
		return
	}
	old, err := s.Alerts.Channels(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	byID := map[string]notify.Channel{}
	for _, c := range old {
		byID[c.ID] = c
	}
	// a channel sent without a secret keeps the one already stored
	chans := make([]notify.Channel, len(in.Channels))
	for i, d := range in.Channels {
		c := d.Channel
		if prev, ok := byID[c.ID]; ok && c.ID != "" {
			if c.Token == "" {
				c.Token = prev.Token
			}
			if c.Password == "" {
				c.Password = prev.Password
			}
		}
		chans[i] = c
	}
	if err := s.Alerts.SaveChannels(r.Context(), chans); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.Alerts.SaveRules(r.Context(), in.Rules); err != nil {
		serverError(w, err)
		return
	}
	s.alertsView(w, r)
}

func (s *Server) testAlert(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ID string `json:"id"`
	}
	if !decode(w, r, &in) {
		return
	}
	chans, err := s.Alerts.Channels(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	for _, c := range chans {
		if c.ID != in.ID {
			continue
		}
		msg := notify.Message{Title: "Atalaia: test", Body: "If you can read this, alerts from Atalaia reach this channel."}
		if err := s.Sender.Send(r.Context(), c, msg); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	writeError(w, http.StatusNotFound, "unknown channel")
}

// localMACs returns the hardware addresses of this machine's interfaces, so the
// device list can mark "this machine".
func localMACs() map[string]bool {
	out := map[string]bool{}
	ifs, err := net.Interfaces()
	if err != nil {
		return out
	}
	for _, i := range ifs {
		if len(i.HardwareAddr) == 6 {
			out[i.HardwareAddr.String()] = true
		}
	}
	return out
}
