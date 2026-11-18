package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MultiIPThreshold is how many addresses one MAC must answer for, in a single
// scan, to be flagged. Routers and hosts with virtual addresses can do this
// legitimately; mark such a device as known to silence the warning.
const MultiIPThreshold = 3

const (
	gatewayMACKey = "gateway_mac"
	multiIPKey    = "multi_ip_macs"
)

// Finding is a security observation made while recording a scan.
type Finding struct {
	Kind   string // EventGatewayChanged or EventMultiIP
	Device Device
	Detail string
}

func txSetting(ctx context.Context, tx *sql.Tx, key string) (string, error) {
	var v string
	err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

func txSetSetting(ctx context.Context, tx *sql.Tx, key, value string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// detectARP looks for the two signs of ARP spoofing that need no privileges:
// the default gateway's address suddenly belonging to another MAC, and one MAC
// answering for many addresses. Each is reported once, when it starts.
func detectARP(ctx context.Context, tx *sql.Tx, obs []Observation, at time.Time, res *Result) error {
	now := at.Unix()
	report := func(mac, kind, detail string) error {
		d, found, err := getByMAC(ctx, tx, mac)
		if err != nil || !found {
			return err
		}
		if err := addEvent(ctx, tx, d.ID, kind, detail, now); err != nil {
			return err
		}
		res.Findings = append(res.Findings, Finding{Kind: kind, Device: d, Detail: detail})
		return nil
	}

	// 1. the gateway's MAC changed
	for _, o := range obs {
		if !o.Gateway {
			continue
		}
		prev, err := txSetting(ctx, tx, gatewayMACKey)
		if err != nil {
			return err
		}
		if prev != "" && prev != o.MAC {
			if err := report(o.MAC, EventGatewayChanged, fmt.Sprintf("%s: %s -> %s", o.IP, prev, o.MAC)); err != nil {
				return err
			}
		}
		if prev != o.MAC {
			if err := txSetSetting(ctx, tx, gatewayMACKey, o.MAC); err != nil {
				return err
			}
		}
		break
	}

	// 2. one MAC answering for many addresses (known devices are trusted)
	flagged := map[string]bool{}
	var current []string
	for _, o := range obs {
		if len(o.IPs) < MultiIPThreshold {
			continue
		}
		d, found, err := getByMAC(ctx, tx, o.MAC)
		if err != nil {
			return err
		}
		if found && d.Known {
			continue
		}
		current = append(current, o.MAC)
	}
	sort.Strings(current)
	prevList, err := txSetting(ctx, tx, multiIPKey)
	if err != nil {
		return err
	}
	for _, m := range strings.Split(prevList, ",") {
		if m != "" {
			flagged[m] = true
		}
	}
	for _, o := range obs {
		isNow := false
		for _, m := range current {
			if m == o.MAC {
				isNow = true
			}
		}
		if isNow && !flagged[o.MAC] {
			ips := make([]string, 0, len(o.IPs))
			for _, ip := range o.IPs {
				ips = append(ips, ip.String())
			}
			sort.Strings(ips)
			if err := report(o.MAC, EventMultiIP, strings.Join(ips, ", ")); err != nil {
				return err
			}
		}
	}
	if joined := strings.Join(current, ","); joined != prevList {
		return txSetSetting(ctx, tx, multiIPKey, joined)
	}
	return nil
}
