package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/discovery"
	"github.com/EttoreCSenatore/atalaia/internal/export"
	"github.com/EttoreCSenatore/atalaia/internal/health"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// scanFunc discovers the devices on the network.
type scanFunc func(ctx context.Context, withNames bool) ([]store.Observation, error)

const timeFormat = "2006-01-02 15:04"

func defaultDB() string {
	if p := os.Getenv("ATALAIA_DB"); p != "" {
		return p
	}
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "atalaia", "atalaia.db")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "atalaia.db"
	}
	return filepath.Join(home, ".local", "share", "atalaia", "atalaia.db")
}

func (a *app) flags(name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.err)
	return fs, fs.String("db", defaultDB(), "database file")
}

func (a *app) cmdScan(ctx context.Context, args []string) int {
	fs, db := a.flags("scan")
	noNames := fs.Bool("no-names", false, "do not look up device names (reverse DNS and mDNS)")
	grace := fs.Duration("offline-after", 10*time.Minute, "mark a device offline after this long without seeing it")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	st, err := store.Open(*db)
	if err != nil {
		return a.fail(err)
	}
	defer st.Close()

	obs, err := a.scan(ctx, !*noNames)
	if err != nil {
		return a.fail(err)
	}
	res, err := st.Record(ctx, a.now(), obs, *grace)
	if err != nil {
		return a.fail(err)
	}
	fmt.Fprintf(a.out, "Found %d devices: %d new, %d back online, %d went offline, %d changed address.\n\n",
		len(obs), len(res.New), len(res.CameOnline), len(res.WentOffline), len(res.IPChanged))
	return a.printDevices(ctx, st, false)
}

func (a *app) cmdDevices(ctx context.Context, args []string) int {
	fs, db := a.flags("devices")
	asJSON := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	st, err := store.Open(*db)
	if err != nil {
		return a.fail(err)
	}
	defer st.Close()
	return a.printDevices(ctx, st, *asJSON)
}

func (a *app) cmdHistory(ctx context.Context, args []string) int {
	fs, db := a.flags("history")
	limit := fs.Int("limit", 30, "number of events")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	st, err := store.Open(*db)
	if err != nil {
		return a.fail(err)
	}
	defer st.Close()
	evs, err := st.Events(ctx, *limit)
	if err != nil {
		return a.fail(err)
	}
	if len(evs) == 0 {
		fmt.Fprintln(a.out, "No history yet. Run `atalaia scan` first.")
		return 0
	}
	w := tabwriter.NewWriter(a.out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "WHEN\tEVENT\tDEVICE\tDETAIL")
	for _, e := range evs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.At.Local().Format(timeFormat), e.Kind, label(e.Name, e.Vendor, e.MAC), e.Detail)
	}
	_ = w.Flush()
	return 0
}

type deviceJSON struct {
	MAC        string `json:"mac"`
	IP         string `json:"ip"`
	Name       string `json:"name"`
	Vendor     string `json:"vendor"`
	Randomized bool   `json:"randomized_mac"`
	Online     bool   `json:"online"`
	FirstSeen  string `json:"first_seen"`
	LastSeen   string `json:"last_seen"`
}

func (a *app) printDevices(ctx context.Context, st *store.Store, asJSON bool) int {
	devs, err := st.Devices(ctx)
	if err != nil {
		return a.fail(err)
	}
	if asJSON {
		out := make([]deviceJSON, 0, len(devs))
		for _, d := range devs {
			out = append(out, deviceJSON{d.MAC, d.IP, d.Name, d.Vendor, d.Randomized, d.Online,
				d.FirstSeen.Format(time.RFC3339), d.LastSeen.Format(time.RFC3339)})
		}
		enc := json.NewEncoder(a.out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return a.fail(err)
		}
		return 0
	}
	if len(devs) == 0 {
		fmt.Fprintln(a.out, "No devices yet. Run `atalaia scan` first.")
		return 0
	}
	w := tabwriter.NewWriter(a.out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "STATE\tIP\tMAC\tVENDOR\tNAME\tLAST SEEN")
	for _, d := range devs {
		state := "offline"
		if d.Online {
			state = "online"
		}
		vendor := dash(d.Vendor)
		if d.Randomized {
			vendor = "(random MAC)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", state, dash(d.IP), d.MAC, vendor, dash(d.Name), d.LastSeen.Local().Format(timeFormat))
	}
	_ = w.Flush()
	return 0
}

func (a *app) fail(err error) int {
	fmt.Fprintln(a.err, "error:", err)
	if errors.Is(err, discovery.ErrUnsupported) {
		return 3
	}
	return 1
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func label(name, vendor, mac string) string {
	switch {
	case name != "":
		return name
	case vendor != "":
		return vendor + " (" + mac + ")"
	}
	return mac
}

// cmdExport prints the inventory as CSV or JSON.
func (a *app) cmdExport(ctx context.Context, args []string) int {
	fs, db := a.flags("export")
	format := fs.String("format", "csv", "csv or json")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *format != "csv" && *format != "json" {
		fmt.Fprintln(a.err, "error: --format must be csv or json")
		return 2
	}
	st, err := store.Open(*db)
	if err != nil {
		return a.fail(err)
	}
	defer st.Close()
	devs, err := st.Devices(ctx)
	if err != nil {
		return a.fail(err)
	}
	rows := export.Rows(devs)
	if *format == "json" {
		err = export.JSON(a.out, rows)
	} else {
		err = export.CSV(a.out, rows)
	}
	if err != nil {
		return a.fail(err)
	}
	return 0
}

// cmdBackup writes a consistent copy of the database (safe while the server runs).
func (a *app) cmdBackup(ctx context.Context, args []string) int {
	fs, db := a.flags("backup")
	heartbeat := fs.String("heartbeat", os.Getenv("ATALAIA_HEARTBEAT_BACKUP"), "ping this URL after the backup, URL/fail when it fails (Healthchecks.io style)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	hb, err := health.ParseHeartbeat(*heartbeat)
	if err != nil {
		return a.fail(err)
	}
	ping := func(ok bool) {
		if err := hb.Ping(ctx, ok); err != nil {
			fmt.Fprintln(a.err, "atalaia:", err)
		}
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(a.err, "usage: atalaia backup [--db file] <new-file.db>")
		return 2
	}
	st, err := store.Open(*db)
	if err != nil {
		ping(false)
		return a.fail(err)
	}
	defer st.Close()
	if err := st.Backup(ctx, fs.Arg(0)); err != nil {
		ping(false)
		return a.fail(err)
	}
	if err := st.SetSetting(ctx, store.LastBackupKey, strconv.FormatInt(a.now().Unix(), 10)); err != nil {
		fmt.Fprintln(a.err, "atalaia: backup written, but its time was not recorded:", err)
	}
	ping(true)
	fmt.Fprintf(a.out, "Backup written to %s (only you can read it: it holds your device list and alert secrets).\n", fs.Arg(0))
	return 0
}
