// Package store keeps the device inventory and its history in a local SQLite
// file. Nothing leaves the machine.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // pure-Go driver: keeps the binary static
)

// Event kinds.
const (
	EventJoined    = "joined"     // first time the device was ever seen
	EventOnline    = "online"     // back after being offline
	EventOffline   = "offline"    // not seen for longer than the grace period
	EventIPChanged = "ip_changed" // same device, new address
)

var migrations = []string{
	`CREATE TABLE devices (
		id         INTEGER PRIMARY KEY,
		mac        TEXT    NOT NULL UNIQUE,
		vendor     TEXT    NOT NULL DEFAULT '',
		randomized INTEGER NOT NULL DEFAULT 0,
		name       TEXT    NOT NULL DEFAULT '',
		ip         TEXT    NOT NULL DEFAULT '',
		online     INTEGER NOT NULL DEFAULT 0,
		first_seen INTEGER NOT NULL,
		last_seen  INTEGER NOT NULL
	);
	CREATE TABLE events (
		id        INTEGER PRIMARY KEY,
		device_id INTEGER NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		kind      TEXT    NOT NULL CHECK (kind IN ('joined','online','offline','ip_changed')),
		detail    TEXT    NOT NULL DEFAULT '',
		at        INTEGER NOT NULL
	);
	CREATE INDEX events_at ON events(at DESC);`,
}

// Store is the device inventory.
type Store struct{ db *sql.DB }

// Device is a known device.
type Device struct {
	ID         int64
	MAC        string
	Vendor     string
	Randomized bool
	Name       string
	IP         string
	Online     bool
	FirstSeen  time.Time
	LastSeen   time.Time
}

// Event is something that happened to a device.
type Event struct {
	At     time.Time
	Kind   string
	Detail string
	MAC    string
	Name   string
	Vendor string
}

// Observation is one device seen by a scan.
type Observation struct {
	MAC        string // lowercase, colon-separated
	IP         netip.Addr
	Vendor     string
	Randomized bool
	Name       string
}

// Result summarizes what a scan changed.
type Result struct {
	New         []Device
	CameOnline  []Device
	WentOffline []Device
	IPChanged   []Device
}

// Open opens (creating it if needed) the database at path. The directory is
// created 0700 and the file 0600: the inventory is personal data.
func Open(path string) (*Store, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o700); err != nil {
		return nil, err
	}
	if f, err := os.OpenFile(abs, os.O_CREATE|os.O_RDWR, 0o600); err == nil {
		_ = f.Close()
	} else {
		return nil, err
	}
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "journal_mode(WAL)")
	u := url.URL{Scheme: "file", Path: abs, RawQuery: q.Encode()}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	var v int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
		return err
	}
	if v > len(migrations) {
		return fmt.Errorf("database is from a newer Atalaia (schema %d, this build knows %d)", v, len(migrations))
	}
	for i := v; i < len(migrations); i++ {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// Record stores one scan taken at time at. A device not seen for longer than
// offlineAfter is marked offline (a grace period avoids flapping when a sleeping
// phone misses one sweep). Empty names and vendors never erase known ones.
func (s *Store) Record(ctx context.Context, at time.Time, obs []Observation, offlineAfter time.Duration) (Result, error) {
	var res Result
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer func() { _ = tx.Rollback() }()
	now := at.Unix()
	seen := map[int64]bool{}

	for _, o := range obs {
		d, found, err := getByMAC(ctx, tx, o.MAC)
		if err != nil {
			return res, err
		}
		ip := o.IP.String()
		if !found {
			r, err := tx.ExecContext(ctx, `INSERT INTO devices (mac, vendor, randomized, name, ip, online, first_seen, last_seen)
				VALUES (?, ?, ?, ?, ?, 1, ?, ?)`, o.MAC, o.Vendor, b2i(o.Randomized), o.Name, ip, now, now)
			if err != nil {
				return res, err
			}
			id, _ := r.LastInsertId()
			if err := addEvent(ctx, tx, id, EventJoined, ip, now); err != nil {
				return res, err
			}
			seen[id] = true
			d = Device{ID: id, MAC: o.MAC, Vendor: o.Vendor, Randomized: o.Randomized, Name: o.Name, IP: ip, Online: true,
				FirstSeen: at, LastSeen: at}
			res.New = append(res.New, d)
			continue
		}
		seen[d.ID] = true
		vendor, name := d.Vendor, d.Name
		if o.Vendor != "" {
			vendor = o.Vendor
		}
		if o.Name != "" {
			name = o.Name
		}
		if _, err := tx.ExecContext(ctx, `UPDATE devices SET vendor=?, randomized=?, name=?, ip=?, online=1, last_seen=? WHERE id=?`,
			vendor, b2i(o.Randomized), name, ip, now, d.ID); err != nil {
			return res, err
		}
		d.Vendor, d.Name = vendor, name
		if !d.Online {
			if err := addEvent(ctx, tx, d.ID, EventOnline, ip, now); err != nil {
				return res, err
			}
			res.CameOnline = append(res.CameOnline, d)
		}
		if d.IP != "" && d.IP != ip {
			if err := addEvent(ctx, tx, d.ID, EventIPChanged, d.IP+" -> "+ip, now); err != nil {
				return res, err
			}
			res.IPChanged = append(res.IPChanged, d)
		}
	}

	// devices that were online but have not been seen for too long
	cutoff := at.Add(-offlineAfter).Unix()
	rows, err := tx.QueryContext(ctx, `SELECT id, mac, vendor, randomized, name, ip, online, first_seen, last_seen
		FROM devices WHERE online=1 AND last_seen < ?`, cutoff)
	if err != nil {
		return res, err
	}
	var gone []Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			_ = rows.Close()
			return res, err
		}
		if !seen[d.ID] {
			gone = append(gone, d)
		}
	}
	if err := rows.Close(); err != nil {
		return res, err
	}
	for _, d := range gone {
		if _, err := tx.ExecContext(ctx, `UPDATE devices SET online=0 WHERE id=?`, d.ID); err != nil {
			return res, err
		}
		if err := addEvent(ctx, tx, d.ID, EventOffline, "", now); err != nil {
			return res, err
		}
		d.Online = false
		res.WentOffline = append(res.WentOffline, d)
	}
	return res, tx.Commit()
}

// Devices returns every known device, online first, then by address.
func (s *Store) Devices(ctx context.Context) ([]Device, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, mac, vendor, randomized, name, ip, online, first_seen, last_seen
		FROM devices ORDER BY online DESC, ip, mac`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	sortByAddr(out)
	return out, rows.Err()
}

// Events returns the latest events, newest first.
func (s *Store) Events(ctx context.Context, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT e.at, e.kind, e.detail, d.mac, d.name, d.vendor
		FROM events e JOIN devices d ON d.id = e.device_id ORDER BY e.at DESC, e.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		var at int64
		if err := rows.Scan(&at, &e.Kind, &e.Detail, &e.MAC, &e.Name, &e.Vendor); err != nil {
			return nil, err
		}
		e.At = time.Unix(at, 0).UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getByMAC(ctx context.Context, q queryer, mac string) (Device, bool, error) {
	row := q.QueryRowContext(ctx, `SELECT id, mac, vendor, randomized, name, ip, online, first_seen, last_seen FROM devices WHERE mac=?`, mac)
	d, err := scanDevice(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Device{}, false, nil
	}
	return d, err == nil, err
}

type scanner interface{ Scan(dest ...any) error }

func scanDevice(r scanner) (Device, error) {
	var d Device
	var rnd, on, first, last int64
	if err := r.Scan(&d.ID, &d.MAC, &d.Vendor, &rnd, &d.Name, &d.IP, &on, &first, &last); err != nil {
		return d, err
	}
	d.Randomized, d.Online = rnd != 0, on != 0
	d.FirstSeen, d.LastSeen = time.Unix(first, 0).UTC(), time.Unix(last, 0).UTC()
	return d, nil
}

func addEvent(ctx context.Context, tx *sql.Tx, id int64, kind, detail string, at int64) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO events (device_id, kind, detail, at) VALUES (?, ?, ?, ?)`, id, kind, detail, at)
	return err
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func sortByAddr(d []Device) {
	less := func(a, b Device) bool {
		if a.Online != b.Online {
			return a.Online
		}
		x, ex := netip.ParseAddr(a.IP)
		y, ey := netip.ParseAddr(b.IP)
		if ex != nil || ey != nil {
			return a.IP < b.IP
		}
		return x.Less(y)
	}
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && less(d[j], d[j-1]); j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}
