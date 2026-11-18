package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/alerts"
	"github.com/EttoreCSenatore/atalaia/internal/api"
	"github.com/EttoreCSenatore/atalaia/internal/auth"
	"github.com/EttoreCSenatore/atalaia/internal/monitor"
	"github.com/EttoreCSenatore/atalaia/internal/notify"
	"github.com/EttoreCSenatore/atalaia/internal/store"
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
	var hosts stringList
	fs.Var(&hosts, "allow-host", "extra host name the app may be opened with (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
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
	srv := &api.Server{Store: st, Monitor: mon, Alerts: svc, Sender: sender, Version: version.Version, AllowedHosts: hosts, Static: web.Handler(), Now: a.now}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		return a.fail(err)
	}
	defer ln.Close()
	url := "http://" + ln.Addr().String()
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
	if host, _, _ := net.SplitHostPort(ln.Addr().String()); host != "127.0.0.1" && host != "::1" {
		fmt.Fprintln(a.err, "warning: serving the local network over plain HTTP; the password and session are not encrypted on the wire.\n"+
			"         Use it on a trusted network only, or put a TLS reverse proxy or a VPN in front.")
	}

	httpSrv := &http.Server{
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    16 << 10,
		// no WriteTimeout: live updates (SSE) keep the response open
	}
	go mon.Run(ctx)
	errc := make(chan error, 1)
	go func() { errc <- httpSrv.Serve(ln) }()
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
