package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/tools"
)

// exposureDTO is the report of an exposure check, built from the device's last
// port scan: risky services, what the web ports answer, and the TLS state.
type exposureDTO struct {
	At    string          `json:"at"`
	Risks []tools.Risk    `json:"risks"`
	Web   []tools.WebInfo `json:"web"`
	TLS   []tools.TLSInfo `json:"tls"`
}

func (s *Server) exposure(w http.ResponseWriter, r *http.Request) {
	hw, err := net.ParseMAC(r.PathValue("mac"))
	if err != nil || len(hw) != 6 {
		writeError(w, http.StatusBadRequest, "invalid MAC address")
		return
	}
	raw, found, err := s.Store.Exposure(r.Context(), hw.String())
	if err != nil {
		serverError(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "no exposure check yet")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(raw)
}

// exposureScan runs the checks on the ports the last port scan found open.
// It sends one HTTP request per web port and one TLS handshake per TLS port,
// so it has the same fences as the port scan: confirmation, scope, one at a
// time, minimum interval, audit log.
func (s *Server) exposureScan(w http.ResponseWriter, r *http.Request) {
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
	last, found, err := s.Store.Ports(r.Context(), d.MAC)
	if err != nil {
		serverError(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusConflict, "run a port scan on this device first")
		return
	}
	now := s.now()
	s.mu.Lock()
	wait := portScanEvery - now.Sub(s.lastScan)
	if wait > 0 {
		s.mu.Unlock()
		writeError(w, http.StatusTooManyRequests, fmt.Sprintf("active checks are limited; try again in %d s", int(wait.Seconds())+1))
		return
	}
	s.lastScan = now
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), portScanLimit)
	defer cancel()
	open := make([]tools.Port, 0, len(last.Ports))
	for _, p := range last.Ports {
		open = append(open, tools.Port{Port: p.Port, Service: p.Service})
	}
	out := exposureDTO{At: now.Format(time.RFC3339), Risks: tools.Assess(open), Web: []tools.WebInfo{}, TLS: []tools.TLSInfo{}}
	if out.Risks == nil {
		out.Risks = []tools.Risk{}
	}
	for _, p := range open {
		if web, secure := tools.IsWebPort(p.Port); web {
			if info, ok := tools.WebProbe(ctx, ip, p.Port, secure, s.Dial); ok {
				out.Web = append(out.Web, info)
			}
		}
		if tools.TLSPorts[p.Port] {
			if info, ok := tools.TLSCheck(ctx, ip, p.Port, s.Dial, now); ok {
				out.TLS = append(out.TLS, info)
			}
		}
	}
	if ctx.Err() != nil {
		writeError(w, http.StatusGatewayTimeout, "the check was interrupted")
		return
	}
	raw, err := json.Marshal(out)
	if err == nil {
		err = s.Store.SaveExposure(r.Context(), d.MAC, now, raw)
	}
	if err != nil {
		serverError(w, err)
		return
	}
	s.audited(r, "exposure", d.MAC, fmt.Sprintf("%s: %d risky, %d web, %d tls", ip, len(out.Risks), len(out.Web), len(out.TLS)))
	writeJSON(w, http.StatusOK, out)
}
