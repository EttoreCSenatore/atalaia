package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/health"
)

type healthDTO struct {
	// Status is "ok", "degraded" (warnings only, HTTP 200) or "error" (HTTP 503).
	Status   string     `json:"status"`
	Version  string     `json:"version"`
	Database string     `json:"database"`
	LastScan *time.Time `json:"last_scan,omitempty"`
	Problems []string   `json:"problems,omitempty"`
	Warnings []string   `json:"warnings,omitempty"`
}

// health answers without login, for uptime monitors (Uptime Kuma and the
// like). It says whether Atalaia works, never anything about the network:
// no devices, addresses or error texts. HTTP 503 when the database does not
// answer, the last scan failed or scans stopped running.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	out := healthDTO{Status: "ok", Version: s.Version, Database: "ok"}
	if err := s.Store.Ping(ctx); err != nil {
		out.Database = "error"
		out.Problems = append(out.Problems, "database")
	}
	st := s.Monitor.Status()
	if !st.LastScan.IsZero() {
		t := st.LastScan.UTC()
		out.LastScan = &t
		if st.LastError != "" {
			out.Problems = append(out.Problems, "last_scan_failed")
		}
		if iv := s.Monitor.Interval; iv > 0 && !st.Scanning && s.now().Sub(st.LastScan) > 3*iv+time.Minute {
			out.Problems = append(out.Problems, "scans_stale")
		}
	}
	if d, err := health.DiskFor(filepath.Dir(s.Store.Path())); err == nil && d.Low {
		out.Warnings = append(out.Warnings, "disk_low")
	}
	code := http.StatusOK
	switch {
	case len(out.Problems) > 0:
		out.Status, code = "error", http.StatusServiceUnavailable
	case len(out.Warnings) > 0:
		out.Status = "degraded"
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, code, out)
}

// metrics serves Prometheus metrics when MetricsToken is set. Only counts,
// times and sizes: never a device, address or name.
func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	if s.MetricsToken == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	want := sha256.Sum256([]byte(s.MetricsToken))
	have := sha256.Sum256([]byte(got))
	if !ok || subtle.ConstantTimeCompare(want[:], have[:]) != 1 {
		w.Header().Set("WWW-Authenticate", `Bearer realm="atalaia-metrics"`)
		writeError(w, http.StatusUnauthorized, "metrics need the bearer token")
		return
	}
	ctx := r.Context()
	var m health.Metrics
	m.Add("atalaia_build_info", "gauge", "Atalaia version (always 1).", 1, "version", s.Version)
	up := 1.0
	if err := s.Store.Ping(ctx); err != nil {
		up = 0
	}
	m.Add("atalaia_database_up", "gauge", "1 when the database answers.", up)
	if total, online, known, err := s.Store.DeviceCounts(ctx); err == nil {
		m.Add("atalaia_devices", "gauge", "Devices in the inventory.", float64(total))
		m.Add("atalaia_devices_online", "gauge", "Devices online at the last scan.", float64(online))
		m.Add("atalaia_devices_known", "gauge", "Devices marked as known.", float64(known))
	}
	st := s.Monitor.Status()
	m.Add("atalaia_scans_total", "counter", "Scans since the server started.", float64(st.Scans))
	m.Add("atalaia_scan_failures_total", "counter", "Failed scans since the server started.", float64(st.Failures))
	if !st.LastScan.IsZero() {
		m.Add("atalaia_last_scan_timestamp_seconds", "gauge", "When the last scan ran (Unix time).", float64(st.LastScan.Unix()))
		m.Add("atalaia_last_scan_duration_seconds", "gauge", "How long the last scan took.", st.Duration)
		ok := 1.0
		if st.LastError != "" {
			ok = 0
		}
		m.Add("atalaia_last_scan_success", "gauge", "1 when the last scan worked.", ok)
	}
	if t, err := s.Store.LastBackup(ctx); err == nil && !t.IsZero() {
		m.Add("atalaia_last_backup_timestamp_seconds", "gauge", "When `atalaia backup` last succeeded (Unix time).", float64(t.Unix()))
	}
	if fi, err := os.Stat(s.Store.Path()); err == nil {
		size := fi.Size()
		if wal, err := os.Stat(s.Store.Path() + "-wal"); err == nil {
			size += wal.Size() // recent writes live in the write-ahead log until a checkpoint
		}
		m.Add("atalaia_database_size_bytes", "gauge", "Size of the database file and its write-ahead log.", float64(size))
	}
	if d, err := health.DiskFor(filepath.Dir(s.Store.Path())); err == nil {
		m.Add("atalaia_disk_free_bytes", "gauge", "Free space on the database's file system.", float64(d.Free))
		m.Add("atalaia_disk_size_bytes", "gauge", "Size of the database's file system.", float64(d.Total))
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(m.String()))
}
