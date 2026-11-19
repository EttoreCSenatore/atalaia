// Command atalaia is the Atalaia network watchtower.
//
// M1 ships device discovery on the command line: `scan`, `devices` and
// `history`. The embedded web app arrives in M2.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/probe"
	"github.com/EttoreCSenatore/atalaia/internal/scope"
	"github.com/EttoreCSenatore/atalaia/internal/version"
)

const usage = `Atalaia: open-source network watchtower.

Usage:
  atalaia serve [--listen ADDR] [--db FILE] [--interval DURATION] [--allow-host NAME]
                [--allow-cidr NET] [--tls-self-signed | --tls-cert F --tls-key F]
                [--no-active-tools] [--secure-cookies]
                    run the web app and scan the network in the background
  atalaia scan [--db FILE] [--no-names] [--offline-after DURATION]
                    discover the devices on the local network and record them
  atalaia devices [--db FILE] [--json]
                    list the known devices
  atalaia history [--db FILE] [--limit N]
                    show what happened (joined, online, offline, new address)
  atalaia export [--db FILE] [--format csv|json]
                    print the device list (CSV is safe to open in a spreadsheet)
  atalaia backup [--db FILE] NEW-FILE
                    write a consistent copy of the database (works while serving)
  atalaia scope     print the networks Atalaia is allowed to probe
  atalaia version   print the version

The database defaults to $ATALAIA_DB or ~/.local/share/atalaia/atalaia.db.
Use Atalaia only on networks you administer or have written authorization for.
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	a := &app{out: os.Stdout, err: os.Stderr, now: time.Now, scan: probe.Live}
	os.Exit(a.run(ctx, os.Args[1:]))
}

type app struct {
	out, err io.Writer
	now      func() time.Time
	scan     scanFunc
}

func (a *app) run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprint(a.err, usage)
		return 2
	}
	rest := args[1:]
	switch args[0] {
	case "serve":
		return a.cmdServe(ctx, rest)
	case "scan":
		return a.cmdScan(ctx, rest)
	case "devices":
		return a.cmdDevices(ctx, rest)
	case "history":
		return a.cmdHistory(ctx, rest)
	case "export":
		return a.cmdExport(ctx, rest)
	case "backup":
		return a.cmdBackup(ctx, rest)
	case "version", "--version":
		fmt.Fprintln(a.out, "atalaia", version.Version)
		return 0
	case "scope":
		s, err := scope.Local()
		if err != nil {
			fmt.Fprintln(a.err, "error:", err)
			return 1
		}
		ps := s.Prefixes()
		if len(ps) == 0 {
			fmt.Fprintln(a.err, "no private IPv4 network found; nothing in scope")
			return 1
		}
		for _, p := range ps {
			fmt.Fprintln(a.out, p)
		}
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(a.out, usage)
		return 0
	default:
		fmt.Fprintf(a.err, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}
