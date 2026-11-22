package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/alerts"
	"github.com/EttoreCSenatore/atalaia/internal/api"
	"github.com/EttoreCSenatore/atalaia/internal/auth"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpprobe"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpsniff"
	"github.com/EttoreCSenatore/atalaia/internal/dhcpwatch"
	"github.com/EttoreCSenatore/atalaia/internal/health"
	"github.com/EttoreCSenatore/atalaia/internal/monitor"
	"github.com/EttoreCSenatore/atalaia/internal/notify"
	"github.com/EttoreCSenatore/atalaia/internal/store"
	"github.com/EttoreCSenatore/atalaia/internal/tlsutil"
	"github.com/EttoreCSenatore/atalaia/internal/version"
	"github.com/EttoreCSenatore/atalaia/internal/web"
)

type stringList []string

func (l *stringList) String() string     { return strings.Join(*l, ",") }
func (l *stringList) Set(v string) error { *l = append(*l, v); return nil }

// cmdServe runs the background scanner and the web app + API.
func (a *app) cmdServe(ctx context.Context, args []string) int {
	fs, db := a.flags("serve")
	listenDefault := os.Getenv("ATALAIA_LISTEN")
	if listenDefault == "" {
		listenDefault = "127.0.0.1:8080"
	}
	listen := fs.String("listen", listenDefault, "address to listen on (use 0.0.0.0:8080 to serve the whole local network)")
	interval := fs.Duration("interval", 5*time.Minute, "time between automatic scans (0 = only when asked)")
	tlsCert := fs.String("tls-cert", os.Getenv("ATALAIA_TLS_CERT"), "serve HTTPS with this certificate file (PEM; needs --tls-key)")
	tlsKey := fs.String("tls-key", os.Getenv("ATALAIA_TLS_KEY"), "private key file for --tls-cert (PEM)")
	tlsSelf := fs.Bool("tls-self-signed", os.Getenv("ATALAIA_TLS_SELF_SIGNED") == "1", "serve HTTPS with a self-signed certificate Atalaia makes and renews (browsers will warn once)")
	secure := fs.Bool("secure-cookies", os.Getenv("ATALAIA_SECURE_COOKIES") == "1", "mark the session cookie Secure even over plain HTTP (use behind a TLS reverse proxy)")
	noActive := fs.Bool("no-active-tools", os.Getenv("ATALAIA_NO_ACTIVE_TOOLS") == "1", "observe only: disable ping, Wake-on-LAN, port scan and the exposure check")
	heartbeat := fs.String("heartbeat-scan", os.Getenv("ATALAIA_HEARTBEAT_SCAN"), "ping this URL after every scan, URL/fail when it fails (Healthchecks.io style; keep it secret)")
	metricsFile := fs.String("metrics-token-file", os.Getenv("ATALAIA_METRICS_TOKEN_FILE"), "enable /metrics for Prometheus with the bearer token in this file (32+ characters; or env ATALAIA_METRICS_TOKEN)")
	var cidrs stringList
	fs.Var(&cidrs, "allow-cidr", "extra client network to treat as local, e.g. 100.64.0.0/10 for Tailscale (repeatable; env ATALAIA_ALLOW_CIDR, comma separated)")
	var hosts stringList
	fs.Var(&hosts, "allow-host", "extra host name the app may be opened with (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if (*tlsCert == "") != (*tlsKey == "") {
		return a.fail(errors.New("--tls-cert and --tls-key must be used together"))
	}
	if *tlsSelf && *tlsCert != "" {
		return a.fail(errors.New("use either --tls-self-signed or --tls-cert/--tls-key, not both"))
	}
	if env := os.Getenv("ATALAIA_ALLOW_CIDR"); env != "" {
		cidrs = append(cidrs, strings.Split(env, ",")...)
	}
	nets, err := api.ParseAllowedNets(cidrs)
	if err != nil {
		return a.fail(err)
	}
	hb, err := health.ParseHeartbeat(*heartbeat)
	if err != nil {
		return a.fail(err)
	}
	metricsToken, err := readMetricsToken(*metricsFile)
	if err != nil {
		return a.fail(err)
	}
	st, err := store.Open(*db)
	if err != nil {
		return a.fail(err)
	}
	defer st.Close()

	offlineAfter := 10 * time.Minute
	if 3**interval > offlineAfter {
		offlineAfter = 3 * *interval
	}
	sender := &notify.Sender{}
	svc := &alerts.Service{Store: st, Sender: sender}
	learner := &dhcpsniff.Learner{Store: st}
	sniffer := &dhcpsniff.Sniffer{Handle: learner.Handle, Listen: func() (net.PacketConn, error) {
		c, err := dhcpprobe.Listen(dhcpprobe.ServerPort)
		if err != nil {
			return nil, err
		}
		return c, nil
	}}
	watcher := &dhcpwatch.Watcher{Store: st, Alerts: svc, Now: a.now, Probe: dhcpProbe}
	var srv *api.Server // set below; the scan hook runs only after the server exists
	mon := &monitor.Monitor{Store: st, Scan: monitor.ScanFunc(a.scan), Interval: *interval, OfflineAfter: offlineAfter, Now: a.now,
		OnResult: func(c context.Context, res store.Result, baseline bool) {
			go svc.Handle(context.WithoutCancel(c), res, baseline)
			go learner.Flush(context.WithoutCancel(c))
			go srv.RunScheduledPortScans(context.WithoutCancel(c))
		}}
	if hb != nil {
		mon.OnDone = func(c context.Context, scanErr error) {
			go func() {
				if err := hb.Ping(context.WithoutCancel(c), scanErr == nil); err != nil {
					fmt.Fprintln(a.err, "atalaia:", err)
				}
			}()
		}
	}
	srv = &api.Server{AllowedNets: nets, NoActiveTools: *noActive, SecureCookies: *secure, Store: st, Monitor: mon, Alerts: svc, DHCP: watcher, Sniff: sniffer, Learner: learner, Sender: sender, Version: version.Version, AllowedHosts: hosts, MetricsToken: metricsToken, Static: web.Handler(), Now: a.now}

	var tlsConf *tls.Config
	if *tlsSelf || *tlsCert != "" {
		cert, key := *tlsCert, *tlsKey
		if *tlsSelf {
			cert, key, err = tlsutil.SelfSigned(filepath.Join(filepath.Dir(*db), "tls"), certNames(hosts), a.now())
			if err != nil {
				return a.fail(err)
			}
			if fp, ferr := tlsutil.Fingerprint(cert); ferr == nil {
				fmt.Fprintf(a.out, "Self-signed certificate (SHA-256): %s\nCompare it with what your browser shows before accepting it.\n", fp)
			}
		}
		tlsConf, err = tlsutil.Config(cert, key)
		if err != nil {
			return a.fail(err)
		}
		srv.HSTS = !*tlsSelf // HSTS makes certificate errors impossible to click through: only with a certificate the user chose
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			return a.fail(fmt.Errorf("%s is already in use by another program; pick another with --listen (for example 127.0.0.1:8090)", *listen))
		}
		return a.fail(err)
	}
	defer ln.Close()
	scheme := "http"
	if tlsConf != nil {
		scheme = "https"
	}
	url := scheme + "://" + ln.Addr().String()
	fmt.Fprintf(a.out, "Atalaia %s listening on %s\n", version.Version, url)

	if has, err := st.HasUser(ctx); err != nil {
		return a.fail(err)
	} else if !has {
		code, err := auth.NewSetupCode()
		if err != nil {
			return a.fail(err)
		}
		srv.SetSetupCode(code)
		fmt.Fprintf(a.out, "\nFirst run: open the address above in a browser and create the administrator account.\nSetup code (shown only here, valid until the account is created): %s\n\n", code)
	}
	if host, _, _ := net.SplitHostPort(ln.Addr().String()); tlsConf == nil && host != "127.0.0.1" && host != "::1" {
		fmt.Fprintln(a.err, "warning: serving the local network over plain HTTP; the password and session are not encrypted on the wire.\n"+
			"         Use it on a trusted network only, or put a TLS reverse proxy or a VPN in front.")
	}

	httpSrv := &http.Server{
		Handler:           srv.Handler(),
		TLSConfig:         tlsConf,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    16 << 10,
		// no WriteTimeout: live updates (SSE) keep the response open
	}
	go mon.Run(ctx)
	go sniffer.Run(ctx, func() bool {
		n, err := st.SettingInt(ctx, api.SniffKey, 0)
		return err == nil && n != 0
	}, 20*time.Second, func(err error) string {
		if dhcpprobe.IsPermission(err) {
			return "binding UDP port 67 needs root or CAP_NET_BIND_SERVICE; see the docs"
		}
		return "could not open UDP port 67 (is a DHCP server running on this machine?)"
	})
	errc := make(chan error, 1)
	go func() {
		if tlsConf != nil {
			errc <- httpSrv.ServeTLS(ln, "", "")
			return
		}
		errc <- httpSrv.Serve(ln)
	}()
	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return a.fail(err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdown)
	}
	return 0
}

// certNames are the names a self-signed certificate must cover: this machine's
// names and local addresses, plus the extra hosts the user allowed.
func certNames(extra []string) []string {
	names := []string{"localhost", "127.0.0.1", "::1"}
	if h, err := os.Hostname(); err == nil && h != "" {
		names = append(names, h)
	}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok && !ipn.IP.IsLinkLocalUnicast() {
				names = append(names, ipn.IP.String())
			}
		}
	}
	return append(names, extra...)
}

// dhcpProbe is the real DHCP check: it needs UDP port 68, a privileged port.
func dhcpProbe(ctx context.Context) ([]dhcpprobe.Offer, error) {
	conn, err := dhcpprobe.Listen(dhcpprobe.ClientPort)
	if err != nil {
		if dhcpprobe.IsPermission(err) {
			return nil, errors.New("binding UDP port 68 needs root or CAP_NET_BIND_SERVICE; see the DHCP section of the docs")
		}
		return nil, errors.New("could not open UDP port 68 (is another DHCP client using it?)")
	}
	defer conn.Close()
	dest := &net.UDPAddr{IP: net.IPv4bcast, Port: dhcpprobe.ServerPort}
	return dhcpprobe.Probe(ctx, conn, dest, 4*time.Second)
}

// readMetricsToken returns the /metrics token from a file (preferred: it does
// not show in the process list), a systemd credential named metrics-token, or
// ATALAIA_METRICS_TOKEN. Empty = /metrics off.
func readMetricsToken(file string) (string, error) {
	tok := os.Getenv("ATALAIA_METRICS_TOKEN")
	if dir := os.Getenv("CREDENTIALS_DIRECTORY"); file == "" && dir != "" {
		if p := filepath.Join(dir, "metrics-token"); fileExists(p) {
			file = p
		}
	}
	if file != "" {
		b, err := os.ReadFile(file) // #nosec G703 -- the path is the operator's own --metrics-token-file flag
		if err != nil {
			return "", fmt.Errorf("metrics token file: %w", err)
		}
		tok = string(b)
	}
	tok = strings.TrimSpace(tok)
	if tok != "" && len(tok) < 32 {
		return "", errors.New("the metrics token must have at least 32 characters (for example: openssl rand -hex 32)")
	}
	return tok, nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p) // #nosec G703 -- a path under systemd's CREDENTIALS_DIRECTORY
	return err == nil
}
