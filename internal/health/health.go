// Package health holds what Atalaia reports about itself: free disk space next
// to the database, heartbeat pings to an external monitor (Healthchecks.io
// style) and the Prometheus text format for /metrics. It never reports
// anything about the devices on the network beyond counts.
package health

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Disk is the space on the file system that holds the database.
type Disk struct {
	Free  uint64 `json:"free_bytes"`
	Total uint64 `json:"total_bytes"`
	Low   bool   `json:"low"`
}

// LowFree and LowPercent decide when the disk is "almost full": less than
// 512 MiB or less than 5% free, whichever comes first.
const (
	LowFree    = 512 << 20
	LowPercent = 5
)

// DiskFor reports the space on the file system holding path.
func DiskFor(path string) (Disk, error) {
	free, total, err := statfs(path)
	if err != nil {
		return Disk{}, err
	}
	return Disk{Free: free, Total: total, Low: IsLow(free, total)}, nil
}

// IsLow reports whether free space is below LowFree or LowPercent of total.
func IsLow(free, total uint64) bool {
	if total == 0 {
		return false
	}
	return free < LowFree || free*100 < total*LowPercent
}

// Heartbeat pings an external monitor after a job: the URL on success and
// URL + "/fail" on failure (the Healthchecks.io convention; other services
// that only know success simply see a missed ping). The URL is a secret: it is
// never logged.
type Heartbeat struct {
	URL    string
	Client *http.Client
}

// ParseHeartbeat validates a heartbeat URL (http or https with a host).
// An empty string means "off".
func ParseHeartbeat(raw string) (*Heartbeat, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return nil, errors.New("heartbeat URL must be an http(s) URL")
	}
	return &Heartbeat{URL: strings.TrimRight(u.String(), "/")}, nil
}

// Ping sends one ping. It waits at most 10 seconds and follows no redirects.
func (h *Heartbeat) Ping(ctx context.Context, ok bool) error {
	if h == nil {
		return nil
	}
	c := h.Client
	if c == nil {
		c = &http.Client{
			Timeout:       10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}
	target := h.URL
	if !ok {
		target += "/fail"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return errors.New("heartbeat: bad URL")
	}
	req.Header.Set("User-Agent", "atalaia-heartbeat")
	resp, err := c.Do(req)
	if err != nil {
		// the error text would contain the secret URL
		return errors.New("heartbeat: request failed")
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("heartbeat: server answered %d", resp.StatusCode)
	}
	return nil
}

// Metrics is a Prometheus text-format writer for gauges and counters.
type Metrics struct{ b strings.Builder }

// Add writes one metric with its HELP and TYPE lines. labels are name/value
// pairs; values are escaped.
func (m *Metrics) Add(name, kind, help string, value float64, labels ...string) {
	fmt.Fprintf(&m.b, "# HELP %s %s\n# TYPE %s %s\n%s", name, help, name, kind, name)
	if len(labels) > 1 {
		m.b.WriteString("{")
		for i := 0; i+1 < len(labels); i += 2 {
			if i > 0 {
				m.b.WriteString(",")
			}
			fmt.Fprintf(&m.b, "%s=%q", labels[i], escapeLabel(labels[i+1]))
		}
		m.b.WriteString("}")
	}
	fmt.Fprintf(&m.b, " %g\n", value)
}

func escapeLabel(v string) string {
	// %q escapes quotes and backslashes; Prometheus also wants newlines as \n,
	// which %q does too. Control characters are dropped.
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' {
			return -1
		}
		return r
	}, v)
}

// String is the exposition text.
func (m *Metrics) String() string { return m.b.String() }
