package store

import (
	"context"
	"time"
)

// DHCPServer is a DHCP server seen on the network.
type DHCPServer struct {
	IP        string
	FirstSeen time.Time
	LastSeen  time.Time
	Detail    string // JSON of its last offer
}

// DHCPObservation is one server that answered a DHCP probe.
type DHCPObservation struct {
	IP     string
	Detail string
}

// DHCPResult says what a probe changed.
type DHCPResult struct {
	// Baseline is true when no server was known before: the first probe only
	// records who the legitimate servers are.
	Baseline bool
	// New are the servers never seen before (empty on a baseline).
	New []DHCPServer
}

// RecordDHCP stores a probe's answers. A server never seen before, when others
// were already known, is returned in New and, if its address belongs to a device
// of the inventory, logged in that device's history as rogue_dhcp.
func (s *Store) RecordDHCP(ctx context.Context, at time.Time, obs []DHCPObservation) (DHCPResult, error) {
	var res DHCPResult
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer func() { _ = tx.Rollback() }()
	var known int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dhcp_servers`).Scan(&known); err != nil {
		return res, err
	}
	res.Baseline = known == 0 && len(obs) > 0
	now := at.Unix()
	for _, o := range obs {
		var seen int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dhcp_servers WHERE ip=?`, o.IP).Scan(&seen); err != nil {
			return res, err
		}
		if seen > 0 {
			if _, err := tx.ExecContext(ctx, `UPDATE dhcp_servers SET last_seen=?, detail=? WHERE ip=?`, now, o.Detail, o.IP); err != nil {
				return res, err
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO dhcp_servers (ip, first_seen, last_seen, detail) VALUES (?, ?, ?, ?)`, o.IP, now, now, o.Detail); err != nil {
			return res, err
		}
		if res.Baseline {
			continue
		}
		res.New = append(res.New, DHCPServer{IP: o.IP, FirstSeen: at, LastSeen: at, Detail: o.Detail})
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM devices WHERE ip=? ORDER BY last_seen DESC LIMIT 1`, o.IP).Scan(&id); err == nil {
			if err := addEvent(ctx, tx, id, EventRogueDHCP, o.IP, now); err != nil {
				return res, err
			}
		}
	}
	return res, tx.Commit()
}

// DHCPServers returns the DHCP servers seen so far, by address.
func (s *Store) DHCPServers(ctx context.Context) ([]DHCPServer, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT ip, first_seen, last_seen, detail FROM dhcp_servers ORDER BY ip`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DHCPServer
	for rows.Next() {
		var d DHCPServer
		var a, b int64
		if err := rows.Scan(&d.IP, &a, &b, &d.Detail); err != nil {
			return nil, err
		}
		d.FirstSeen, d.LastSeen = time.Unix(a, 0).UTC(), time.Unix(b, 0).UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

// ForgetDHCPServers clears the remembered servers, so the next probe is a new baseline
// (after replacing the router, for instance).
func (s *Store) ForgetDHCPServers(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM dhcp_servers`)
	return err
}
