// Package export writes the device inventory as CSV or JSON, for spreadsheets,
// backups of the list or other tools.
package export

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
)

// Row is one device in an export.
type Row struct {
	MAC       string   `json:"mac"`
	IP        string   `json:"ip"`
	Name      string   `json:"name"`
	Label     string   `json:"label"`
	Vendor    string   `json:"vendor"`
	Model     string   `json:"model"`
	Kind      string   `json:"kind"`
	Tags      []string `json:"tags"`
	Known     bool     `json:"known"`
	Online    bool     `json:"online"`
	RandomMAC bool     `json:"randomized_mac"`
	FirstSeen string   `json:"first_seen"`
	LastSeen  string   `json:"last_seen"`
}

// Rows converts devices for export.
func Rows(devs []store.Device) []Row {
	out := make([]Row, 0, len(devs))
	for _, d := range devs {
		tags := d.Tags
		if tags == nil {
			tags = []string{}
		}
		out = append(out, Row{d.MAC, d.IP, d.Name, d.Label, d.Vendor, d.Model, d.Kind, tags, d.Known, d.Online, d.Randomized,
			d.FirstSeen.Format(time.RFC3339), d.LastSeen.Format(time.RFC3339)})
	}
	return out
}

// JSON writes the rows as a JSON array.
func JSON(w io.Writer, rows []Row) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

// CSV writes the rows with a header. Cells that a spreadsheet would run as a
// formula (starting with = + - @, tab or carriage return) are prefixed with a
// quote: device names come from the network and must not become formulas.
func CSV(w io.Writer, rows []Row) error {
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"mac", "ip", "name", "label", "vendor", "model", "kind", "tags", "known", "online", "randomized_mac", "first_seen", "last_seen"})
	b := func(v bool) string {
		if v {
			return "yes"
		}
		return "no"
	}
	for _, r := range rows {
		rec := []string{r.MAC, r.IP, r.Name, r.Label, r.Vendor, r.Model, r.Kind, strings.Join(r.Tags, " "), b(r.Known), b(r.Online), b(r.RandomMAC), r.FirstSeen, r.LastSeen}
		for i, c := range rec {
			rec[i] = safeCell(c)
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func safeCell(c string) string {
	if c != "" && strings.ContainsRune("=+-@\t\r", rune(c[0])) {
		return "'" + c
	}
	return c
}
