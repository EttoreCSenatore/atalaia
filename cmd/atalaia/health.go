package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// cmdHealth asks a running server for /api/health and exits 0 when it is
// healthy: for Docker's HEALTHCHECK and scripts (the image has no curl).
func (a *app) cmdHealth(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	fs.SetOutput(a.err)
	target := fs.String("url", defaultHealthURL(), "server to check")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	u, err := url.Parse(*target)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		fmt.Fprintln(a.err, "atalaia: --url must be like http://127.0.0.1:8080")
		return 2
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if ip := net.ParseIP(u.Hostname()); ip != nil && ip.IsLoopback() {
		// this machine's own server, usually with the self-signed certificate
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- loopback only
	}
	c := &http.Client{Timeout: 5 * time.Second, Transport: tr,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(u.String(), "/")+"/api/health", nil)
	if err != nil {
		fmt.Fprintln(a.err, "atalaia:", err)
		return 2
	}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Fprintln(a.err, "atalaia: server not reachable:", err)
		return 1
	}
	defer resp.Body.Close()
	var h struct {
		Status   string   `json:"status"`
		Problems []string `json:"problems"`
		Warnings []string `json:"warnings"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&h)
	line := h.Status
	if len(h.Problems) > 0 {
		line += " (" + strings.Join(h.Problems, ", ") + ")"
	}
	if len(h.Warnings) > 0 {
		line += " [" + strings.Join(h.Warnings, ", ") + "]"
	}
	fmt.Fprintln(a.out, line)
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

// defaultHealthURL is the server this machine runs, from the same
// environment variables `atalaia serve` reads.
func defaultHealthURL() string {
	listen := os.Getenv("ATALAIA_LISTEN")
	if listen == "" {
		listen = "127.0.0.1:8080"
	}
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		host, port = "127.0.0.1", "8080"
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		host = "127.0.0.1"
	}
	scheme := "http"
	if os.Getenv("ATALAIA_TLS_SELF_SIGNED") == "1" || os.Getenv("ATALAIA_TLS_CERT") != "" {
		scheme = "https"
	}
	return scheme + "://" + net.JoinHostPort(host, port)
}
