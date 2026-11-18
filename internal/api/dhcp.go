package api

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/dhcpprobe"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpwatch"
)

type dhcpServerDTO struct {
	IP        string   `json:"ip"`
	FirstSeen string   `json:"first_seen"`
	LastSeen  string   `json:"last_seen"`
	Router    string   `json:"router"`
	DNS       []string `json:"dns"`
}

type sniffDTO struct {
	Enabled bool   `json:"enabled"`
	Active  bool   `json:"active"`
	Error   string `json:"error,omitempty"`
	Learned int    `json:"learned"`
}

type dhcpDTO struct {
	Names   sniffDTO          `json:"names"`
	Enabled bool              `json:"enabled"`
	Servers []dhcpServerDTO   `json:"servers"`
	Last    *dhcpwatch.Report `json:"last"`
}

func (s *Server) dhcpView(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.DHCPServers(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	out := dhcpDTO{Names: s.sniffView(r), Enabled: s.DHCP.Enabled(r.Context()), Servers: make([]dhcpServerDTO, 0, len(list)), Last: s.DHCP.Last()}
	for _, sv := range list {
		var o dhcpprobe.Offer
		_ = json.Unmarshal([]byte(sv.Detail), &o)
		d := dhcpServerDTO{IP: sv.IP, FirstSeen: sv.FirstSeen.Format(time.RFC3339), LastSeen: sv.LastSeen.Format(time.RFC3339), DNS: []string{}}
		if o.Router.IsValid() {
			d.Router = o.Router.String()
		}
		for _, a := range o.DNS {
			d.DNS = append(d.DNS, netip.Addr.String(a))
		}
		out.Servers = append(out.Servers, d)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getDHCP(w http.ResponseWriter, r *http.Request) { s.dhcpView(w, r) }

func (s *Server) putDHCP(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if !decode(w, r, &in) {
		return
	}
	v := "0"
	if in.Enabled {
		v = "1"
	}
	if err := s.Store.SetSetting(r.Context(), dhcpwatch.EnabledKey, v); err != nil {
		serverError(w, err)
		return
	}
	s.audited(r, "dhcp-setting", "", "scheduled check "+map[bool]string{true: "on", false: "off"}[in.Enabled])
	s.dhcpView(w, r)
}

func (s *Server) checkDHCP(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if !decode(w, r, &in) {
		return
	}
	if !in.Confirm {
		writeError(w, http.StatusBadRequest, "confirm that this is your own network (confirm: true)")
		return
	}
	if !s.gate(w) {
		return
	}
	defer s.active.Unlock()
	rep, err := s.DHCP.Check(r.Context())
	s.audited(r, "dhcp", "", summarizeDHCP(rep, err))
	if err == dhcpwatch.ErrBusy {
		writeError(w, http.StatusTooManyRequests, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, rep.Error)
		return
	}
	s.dhcpView(w, r)
}

func summarizeDHCP(rep dhcpwatch.Report, err error) string {
	if err != nil {
		return "failed"
	}
	return itoa(len(rep.Offers)) + " server(s), " + itoa(len(rep.New)) + " new"
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (s *Server) resetDHCP(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if !decode(w, r, &in) {
		return
	}
	if !in.Confirm {
		writeError(w, http.StatusBadRequest, "confirm to forget the known DHCP servers (confirm: true)")
		return
	}
	if err := s.Store.ForgetDHCPServers(r.Context()); err != nil {
		serverError(w, err)
		return
	}
	s.audited(r, "dhcp-reset", "", "known servers forgotten")
	s.dhcpView(w, r)
}

// SniffKey is the setting that turns on learning names from DHCP requests.
const SniffKey = "dhcp_sniff"

func (s *Server) sniffView(r *http.Request) sniffDTO {
	n, _ := s.Store.SettingInt(r.Context(), SniffKey, 0)
	out := sniffDTO{Enabled: n != 0}
	if s.Sniff != nil {
		out.Active, out.Error = s.Sniff.Status()
	}
	if s.Learner != nil {
		out.Learned = s.Learner.Learned
	}
	return out
}

// putSniff turns the passive listener on or off. It only listens (no packet is
// sent), so it is not an active tool; it does need to bind UDP port 67.
func (s *Server) putSniff(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if !decode(w, r, &in) {
		return
	}
	v := "0"
	if in.Enabled {
		v = "1"
	}
	if err := s.Store.SetSetting(r.Context(), SniffKey, v); err != nil {
		serverError(w, err)
		return
	}
	s.audited(r, "dhcp-names", "", "listening "+map[bool]string{true: "on", false: "off"}[in.Enabled])
	s.dhcpView(w, r)
}
