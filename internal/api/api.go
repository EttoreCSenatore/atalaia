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
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/auth"
	"github.com/EttoreCSenatore/atalaia/internal/monitor"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

const (
	cookieName = "atalaia_session"
	sessionTTL = 7 * 24 * time.Hour
	// CSRFHeader must be present on every state-changing request.
	CSRFHeader = "X-Atalaia"
	maxBody    = 4 << 10
	// maxStream bounds a live-update connection; the browser reconnects and
	// the session is checked again.
	maxStream = 30 * time.Minute
)

// Server is the HTTP API.
type Server struct {
	Store   *store.Store
	Monitor *monitor.Monitor
	Version string
	// AllowedHosts are extra Host names accepted besides IP addresses,
	// localhost and the machine's own host name.
	AllowedHosts []string
	// Static serves the web app for every path outside /api/ (may be nil).
	Static http.Handler
	Now    func() time.Time

	limiter auth.Limiter

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

func (s *Server) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Handler returns the full handler with all protections applied.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.status)
	mux.HandleFunc("POST /api/setup", s.mutating(s.setup))
	mux.HandleFunc("POST /api/login", s.mutating(s.login))
	mux.HandleFunc("POST /api/logout", s.mutating(s.protected(s.logout)))
	mux.HandleFunc("POST /api/password", s.mutating(s.protected(s.changePassword)))
	mux.HandleFunc("GET /api/devices", s.protected(s.devices))
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
		if !isLocalClient(r.RemoteAddr) {
			writeError(w, http.StatusForbidden, "this server only answers to the local network")
			return
		}
		next.ServeHTTP(w, r)
	})
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
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "send application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
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
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: token, Path: "/", MaxAge: int(sessionTTL.Seconds()),
		HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil,
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
	out := statusDTO{Version: s.Version, SetupRequired: !has, Authenticated: s.authenticated(r)}
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
	if ok, wait := s.limiter.Allow(key, s.now()); !ok {
		tooMany(w, wait)
		return
	}
	s.mu.Lock()
	code := s.setupCode
	s.mu.Unlock()
	if code == "" || !auth.EqualToken(strings.ToUpper(strings.TrimSpace(in.Code)), code) {
		s.limiter.Fail(key, s.now())
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
	if err := s.startSession(w, r); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
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
	if ok, wait := s.limiter.Allow(key, s.now()); !ok {
		tooMany(w, wait)
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
	if len(in.Password) > 1024 || !auth.VerifyPassword(hash, in.Password) {
		s.limiter.Fail(key, s.now())
		writeError(w, http.StatusUnauthorized, "wrong password")
		return
	}
	s.limiter.Reset(key)
	if err := s.startSession(w, r); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	_ = s.Store.DeleteSession(r.Context(), s.sessionHash(r))
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
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
	if ok, wait := s.limiter.Allow(key, s.now()); !ok {
		tooMany(w, wait)
		return
	}
	hash, err := s.Store.PasswordHash(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	if len(in.Current) > 1024 || !auth.VerifyPassword(hash, in.Current) {
		s.limiter.Fail(key, s.now())
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
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- handlers: data ---------------------------------------------------------

type deviceDTO struct {
	MAC        string `json:"mac"`
	IP         string `json:"ip"`
	Name       string `json:"name"`
	Vendor     string `json:"vendor"`
	Randomized bool   `json:"randomized_mac"`
	Online     bool   `json:"online"`
	FirstSeen  string `json:"first_seen"`
	LastSeen   string `json:"last_seen"`
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
	out := make([]deviceDTO, 0, len(devs))
	for _, d := range devs {
		out = append(out, deviceDTO{d.MAC, d.IP, d.Name, d.Vendor, d.Randomized, d.Online,
			d.FirstSeen.Format(time.RFC3339), d.LastSeen.Format(time.RFC3339)})
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
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	days, err := s.Store.SettingInt(r.Context(), monitor.RetentionKey, monitor.DefaultRetentionDays)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settingsDTO{RetentionDays: days, ScanInterval: s.Monitor.Status().Interval})
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RetentionDays int `json:"retention_days"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.RetentionDays < 0 || in.RetentionDays > 3650 {
		writeError(w, http.StatusBadRequest, "retention_days must be between 0 (keep forever) and 3650")
		return
	}
	if err := s.Store.SetSetting(r.Context(), monitor.RetentionKey, fmt.Sprint(in.RetentionDays)); err != nil {
		serverError(w, err)
		return
	}
	s.getSettings(w, r)
}
