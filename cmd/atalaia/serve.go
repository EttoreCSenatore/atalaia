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
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/alerts"
	"github.com/EttoreCSenatore/atalaia/internal/api"
	"github.com/EttoreCSenatore/atalaia/internal/auth"
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
	listen := fs.String("listen", "127.0.0.1:8080", "address to listen on (use 0.0.0.0:8080 to serve the whole local network)")
	interval := fs.Duration("interval", 5*time.Minute, "time between automatic scans (0 = only when asked)")
	tlsCert := fs.String("tls-cert", os.Getenv("ATALAIA_TLS_CERT"), "serve HTTPS with this certificate file (PEM; needs --tls-key)")
	tlsKey := fs.String("tls-key", os.Getenv("ATALAIA_TLS_KEY"), "private key file for --tls-cert (PEM)")
	tlsSelf := fs.Bool("tls-self-signed", os.Getenv("ATALAIA_TLS_SELF_SIGNED") == "1", "serve HTTPS with a self-signed certificate Atalaia makes and renews (browsers will warn once)")
	secure := fs.Bool("secure-cookies", os.Getenv("ATALAIA_SECURE_COOKIES") == "1", "mark the session cookie Secure even over plain HTTP (use behind a TLS reverse proxy)")
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
	mon := &monitor.Monitor{Store: st, Scan: monitor.ScanFunc(a.scan), Interval: *interval, OfflineAfter: offlineAfter, Now: a.now,
		OnResult: func(c context.Context, res store.Result, baseline bool) {
			go svc.Handle(context.WithoutCancel(c), res, baseline)
		}}
	srv := &api.Server{SecureCookies: *secure, Store: st, Monitor: mon, Alerts: svc, Sender: sender, Version: version.Version, AllowedHosts: hosts, Static: web.Handler(), Now: a.now}

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
