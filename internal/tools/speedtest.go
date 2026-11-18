package tools

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultSpeedURL is the file the speed test downloads. Using it tells that
// server your public IP address, so the app asks for confirmation first.
const DefaultSpeedURL = "https://speed.cloudflare.com/__down?bytes=25000000"

// SpeedResult is the outcome of a download speed test.
type SpeedResult struct {
	Server    string  `json:"server"`     // host contacted
	Mbps      float64 `json:"mbps"`       // download speed, megabits per second
	Bytes     int64   `json:"bytes"`      // bytes downloaded
	Seconds   float64 `json:"seconds"`    // time spent downloading
	LatencyMs float64 `json:"latency_ms"` // time to the first byte
}

// SpeedTest downloads rawURL (https only) for at most maxDur or maxBytes and
// reports the throughput. Redirects are not followed.
func SpeedTest(ctx context.Context, client *http.Client, rawURL string, maxBytes int64, maxDur time.Duration) (SpeedResult, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return SpeedResult{}, errors.New("invalid speed test address")
	}
	res := SpeedResult{Server: u.Hostname()}
	if client == nil {
		client = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	ctx, cancel := context.WithTimeout(ctx, maxDur)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return res, errors.New("invalid speed test address")
	}
	req.Header.Set("User-Agent", "atalaia")
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return res, errors.New("could not reach the speed test server")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return res, errors.New("the speed test server did not answer with a file")
	}
	res.LatencyMs = float64(time.Since(start).Microseconds()) / 1000
	bodyStart := time.Now()
	n, _ := io.Copy(io.Discard, io.LimitReader(resp.Body, maxBytes)) // a timeout simply ends the measurement
	res.Bytes = n
	res.Seconds = time.Since(bodyStart).Seconds()
	if n < 1024 || res.Seconds <= 0 {
		return res, errors.New("not enough data was received to measure")
	}
	res.Mbps = float64(n) * 8 / res.Seconds / 1e6
	return res, nil
}
