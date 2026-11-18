package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// AuditEntry records an active action (port scan, Wake-on-LAN, ping…) taken by the administrator.
type AuditEntry struct {
	At     time.Time
	Action string
	Target string
	Detail string
}

// Audit appends an entry to the audit log.
func (s *Store) Audit(ctx context.Context, at time.Time, action, target, detail string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit (at, action, target, detail) VALUES (?, ?, ?, ?)`,
		at.Unix(), action, target, detail)
	return err
}

// Audits returns the latest entries, newest first.
func (s *Store) Audits(ctx context.Context, limit int) ([]AuditEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT at, action, target, detail FROM audit ORDER BY at DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var at int64
		if err := rows.Scan(&at, &e.Action, &e.Target, &e.Detail); err != nil {
			return nil, err
		}
		e.At = time.Unix(at, 0).UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

// PortResult is the last port scan of a device.
type PortResult struct {
	At    time.Time
	Ports []OpenPort
}

// OpenPort is one open TCP port.
type OpenPort struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
}

// SavePorts stores the latest port scan of the device with that MAC.
func (s *Store) SavePorts(ctx context.Context, mac string, at time.Time, ports []OpenPort) error {
	if ports == nil {
		ports = []OpenPort{}
	}
	b, err := json.Marshal(ports)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO port_results (device_id, at, ports)
		SELECT id, ?, ? FROM devices WHERE mac=?
		ON CONFLICT(device_id) DO UPDATE SET at=excluded.at, ports=excluded.ports`, at.Unix(), string(b), mac)
	return err
}

// Ports returns the last port scan of the device, or false if it was never scanned.
func (s *Store) Ports(ctx context.Context, mac string) (PortResult, bool, error) {
	var at int64
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT p.at, p.ports FROM port_results p JOIN devices d ON d.id = p.device_id WHERE d.mac=?`, mac).Scan(&at, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return PortResult{}, false, nil
	}
	if err != nil {
		return PortResult{}, false, err
	}
	var ports []OpenPort
	if err := json.Unmarshal([]byte(raw), &ports); err != nil {
		return PortResult{}, false, err
	}
	return PortResult{At: time.Unix(at, 0).UTC(), Ports: ports}, true, nil
}

// SaveExposure stores the latest exposure report (JSON) of the device with that MAC.
func (s *Store) SaveExposure(ctx context.Context, mac string, at time.Time, report []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO exposure_results (device_id, at, report)
		SELECT id, ?, ? FROM devices WHERE mac=?
		ON CONFLICT(device_id) DO UPDATE SET at=excluded.at, report=excluded.report`, at.Unix(), string(report), mac)
	return err
}

// Exposure returns the latest exposure report (JSON), or false if there is none.
func (s *Store) Exposure(ctx context.Context, mac string) ([]byte, bool, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT e.report FROM exposure_results e JOIN devices d ON d.id = e.device_id WHERE d.mac=?`, mac).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return []byte(raw), err == nil, err
}
