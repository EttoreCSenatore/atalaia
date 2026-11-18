package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
	"github.com/EttoreCSenatore/atalaia/internal/store"
	"github.com/EttoreCSenatore/atalaia/internal/tools"
)

// Active tools act on the network, so they are fenced in: the signed-in
// administrator only, one at a time, port scans and Wake-on-LAN only on
// devices already in the inventory whose address is in the scan scope, port
// scans need an explicit confirmation and a minimum interval, and every use is
// written to the audit log.
const (
	portScanEvery = 10 * time.Second
	portScanLimit = 60 * time.Second
)

func (s *Server) scopeNow() (*scope.Scope, error) {
	if s.Scope != nil {
		return s.Scope()
	}
	return scope.Local()
}

// activeOnly refuses the active tools when the server runs in observe-only mode.
func (s *Server) activeOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.NoActiveTools {
			writeError(w, http.StatusForbidden, "active tools are disabled on this server (observe-only mode)")
			return
		}
		next(w, r)
	}
}

// gate admits one active tool at a time.
func (s *Server) gate(w http.ResponseWriter) bool {
	if !s.active.TryLock() {
		writeError(w, http.StatusTooManyRequests, "another tool is running; wait a moment")
		return false
	}
	return true
}

func (s *Server) audited(r *http.Request, action, target, detail string) {
	// a failing audit write must not hide the result of an action that already ran
	_ = s.Store.Audit(context.WithoutCancel(r.Context()), s.now(), action, target, detail)
}

// inScopeDevice finds the device named in the path and checks its address is in scope.
func (s *Server) inScopeDevice(w http.ResponseWriter, r *http.Request) (store.Device, netip.Addr, *scope.Scope, bool) {
	hw, err := net.ParseMAC(r.PathValue("mac"))
	if err != nil || len(hw) != 6 {
		writeError(w, http.StatusBadRequest, "invalid MAC address")
		return store.Device{}, netip.Addr{}, nil, false
	}
	devs, err := s.Store.Devices(r.Context())
	if err != nil {
		serverError(w, err)
		return store.Device{}, netip.Addr{}, nil, false
	}
	for _, d := range devs {
		if d.MAC != hw.String() {
			continue
		}
		ip, err := netip.ParseAddr(d.IP)
		sc, serr := s.scopeNow()
		if err != nil || serr != nil || !sc.Allows(ip.Unmap()) {
			writeError(w, http.StatusForbidden, "this device's address is outside the networks Atalaia is allowed to act on")
			return store.Device{}, netip.Addr{}, nil, false
		}
		return d, ip.Unmap(), sc, true
	}
	writeError(w, http.StatusNotFound, "unknown device")
	return store.Device{}, netip.Addr{}, nil, false
}

func (s *Server) wake(w http.ResponseWriter, r *http.Request) {
	d, ip, sc, ok := s.inScopeDevice(w, r)
	if !ok || !s.gate(w) {
		return
	}
	defer s.active.Unlock()
	var bc netip.Addr
	for _, p := range sc.Prefixes() {
		if p.Contains(ip) {
			bc = tools.Broadcast(p)
		}
	}
	hw, _ := net.ParseMAC(d.MAC)
	if !bc.IsValid() {
		writeError(w, http.StatusForbidden, "no local network contains this device")
		return
	}
	if err := tools.Wake(hw, bc); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	s.audited(r, "wake", d.MAC, "magic packet to "+bc.String())
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

type portsDTO struct {
	At     string           `json:"at"`
	Ports  []store.OpenPort `json:"ports"`
	Opened []int            `json:"opened,omitempty"`
	Closed []int            `json:"closed,omitempty"`
}

func newPortsDTO(r store.PortResult) portsDTO {
	p := r.Ports
	if p == nil {
		p = []store.OpenPort{}
	}
	return portsDTO{At: r.At.Format(time.RFC3339), Ports: p}
}

func (s *Server) ports(w http.ResponseWriter, r *http.Request) {
	hw, err := net.ParseMAC(r.PathValue("mac"))
	if err != nil || len(hw) != 6 {
		writeError(w, http.StatusBadRequest, "invalid MAC address")
		return
	}
	res, found, err := s.Store.Ports(r.Context(), hw.String())
	if err != nil {
		serverError(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "this device was never scanned")
		return
	}
	writeJSON(w, http.StatusOK, newPortsDTO(res))
}

func (s *Server) portScan(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if !decode(w, r, &in) {
		return
	}
	if !in.Confirm {
		writeError(w, http.StatusBadRequest, "confirm that this is your own network and device (confirm: true)")
		return
	}
	d, ip, _, ok := s.inScopeDevice(w, r)
	if !ok || !s.gate(w) {
		return
	}
	defer s.active.Unlock()
	now := s.now()
	s.mu.Lock()
	wait := portScanEvery - now.Sub(s.lastScan)
	if wait > 0 {
		s.mu.Unlock()
		writeError(w, http.StatusTooManyRequests, fmt.Sprintf("port scans are limited; try again in %d s", int(wait.Seconds())+1))
		return
	}
	s.lastScan = now
	s.mu.Unlock()

	prev, hadPrev, err := s.Store.Ports(r.Context(), d.MAC)
	if err != nil {
		serverError(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), portScanLimit)
	defer cancel()
	found := tools.PortScan(ctx, ip, s.Dial, 32, 700*time.Millisecond)
	if ctx.Err() != nil {
		writeError(w, http.StatusGatewayTimeout, "the scan was interrupted")
		return
	}
	open := make([]store.OpenPort, 0, len(found))
	for _, p := range found {
		open = append(open, store.OpenPort{Port: p.Port, Service: p.Service})
	}
	if err := s.Store.SavePorts(r.Context(), d.MAC, now, open); err != nil {
		serverError(w, err)
		return
	}
	s.audited(r, "portscan", d.MAC, fmt.Sprintf("%s: %d open", ip, len(open)))
	out := newPortsDTO(store.PortResult{At: now, Ports: open})
	if hadPrev {
		out.Opened, out.Closed = diffPorts(prev.Ports, open)
		if len(out.Opened) > 0 && s.Alerts != nil {
			go s.Alerts.HandlePorts(context.WithoutCancel(r.Context()), d, out.Opened)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func diffPorts(before, after []store.OpenPort) (opened, closed []int) {
	was, is := map[int]bool{}, map[int]bool{}
	for _, p := range before {
		was[p.Port] = true
	}
	for _, p := range after {
		is[p.Port] = true
		if !was[p.Port] {
			opened = append(opened, p.Port)
		}
	}
	for _, p := range before {
		if !is[p.Port] {
			closed = append(closed, p.Port)
		}
	}
	return
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Host  string `json:"host"`
		Count int    `json:"count"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Count < 0 || in.Count > 10 {
		writeError(w, http.StatusBadRequest, "count must be between 1 and 10")
		return
	}
	if !s.gate(w) {
		return
	}
	defer s.active.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	ip, err := tools.ResolveIPv4(ctx, in.Host)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := tools.Ping(ctx, ip, in.Count, time.Second)
	if errors.Is(err, tools.ErrPingUnavailable) {
		writeError(w, http.StatusNotImplemented, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusGatewayTimeout, "the ping was interrupted")
		return
	}
	s.audited(r, "ping", ip.String(), fmt.Sprintf("%d sent, %d lost", res.Sent, res.Lost))
	if res.Replies == nil {
		res.Replies = []tools.PingReply{}
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) dnsLookup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if !decode(w, r, &in) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	ans, err := tools.DNSLookup(ctx, nil, in.Name, in.Type)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": in.Name, "type": in.Type, "answers": ans})
}

type auditDTO struct {
	At     string `json:"at"`
	Action string `json:"action"`
	Target string `json:"target"`
	Detail string `json:"detail"`
}

func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	es, err := s.Store.Audits(r.Context(), 100)
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]auditDTO, 0, len(es))
	for _, e := range es {
		out = append(out, auditDTO{e.At.Format(time.RFC3339), e.Action, e.Target, e.Detail})
	}
	writeJSON(w, http.StatusOK, out)
}
