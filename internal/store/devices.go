package store

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Limits for what the user can attach to a device.
const (
	MaxLabelRunes = 64
	MaxTags       = 8
	MaxTagRunes   = 24
)

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// CleanTags trims, lower-cases and de-duplicates tags and checks the limits.
// A tag may not contain commas (they separate tags in storage) or control characters.
func CleanTags(in []string) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	for _, t := range in {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" || seen[t] {
			continue
		}
		if utf8.RuneCountInString(t) > MaxTagRunes {
			return nil, fmt.Errorf("a tag may have at most %d characters", MaxTagRunes)
		}
		for _, r := range t {
			if r == ',' || r < 0x20 || r == 0x7f {
				return nil, fmt.Errorf("tags cannot contain commas or control characters")
			}
		}
		seen[t] = true
		out = append(out, t)
	}
	if len(out) > MaxTags {
		return nil, fmt.Errorf("at most %d tags per device", MaxTags)
	}
	return out, nil
}

// CleanLabel trims the label and checks it.
func CleanLabel(l string) (string, error) {
	l = strings.TrimSpace(l)
	if utf8.RuneCountInString(l) > MaxLabelRunes {
		return "", fmt.Errorf("the name may have at most %d characters", MaxLabelRunes)
	}
	for _, r := range l {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("the name cannot contain control characters")
		}
	}
	return l, nil
}

// UpdateDevice sets what the user knows about a device. It reports false when
// no device has that address. Inputs must already be cleaned.
func (s *Store) UpdateDevice(ctx context.Context, mac, label string, tags []string, known bool) (Device, bool, error) {
	r, err := s.db.ExecContext(ctx, `UPDATE devices SET label=?, tags=?, known=? WHERE mac=?`,
		label, strings.Join(tags, ","), b2i(known), mac)
	if err != nil {
		return Device{}, false, err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return Device{}, false, nil
	}
	return getByMAC(ctx, s.db, mac)
}

// CountDevices returns how many devices are known.
func (s *Store) CountDevices(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&n)
	return n, err
}
