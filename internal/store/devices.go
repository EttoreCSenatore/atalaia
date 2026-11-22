package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// DeviceCounts returns how many devices are in the inventory, online and marked known.
func (s *Store) DeviceCounts(ctx context.Context) (total, online, known int, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(online), 0), COALESCE(SUM(known), 0) FROM devices`).Scan(&total, &online, &known)
	return
}

// DeviceByMAC returns the device with that MAC, if it is in the inventory.
func (s *Store) DeviceByMAC(ctx context.Context, mac string) (Device, bool, error) {
	return getByMAC(ctx, s.db, mac)
}

// UpdateIdentity fills in what a device announced: the name only when it has
// none, the kind only when it has none or only the generic "mobile" guess. It
// reports whether anything changed. Names and kinds already known are kept.
func (s *Store) UpdateIdentity(ctx context.Context, mac, name, kind string) (bool, error) {
	r, err := s.db.ExecContext(ctx, `UPDATE devices SET
		name = CASE WHEN name = '' THEN ? ELSE name END,
		kind = CASE WHEN kind IN ('', 'mobile') AND ? != '' THEN ? ELSE kind END
		WHERE mac = ? AND ((name = '' AND ? != '') OR (kind IN ('', 'mobile') AND ? != '' AND kind != ?))`,
		name, kind, kind, mac, name, kind, kind)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	return n > 0, nil
}

// DeleteDevice forgets a device and everything about it (history, port and
// exposure results). It reports false when no device has that MAC.
func (s *Store) DeleteDevice(ctx context.Context, mac string) (bool, error) {
	r, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE mac=?`, mac)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	return n > 0, nil
}

// PruneDevices forgets devices that have been offline since before `before`
// and that the user never touched (not known, no name, no tags): phones that
// rotate their random MAC otherwise pile up forever. It returns how many.
func (s *Store) PruneDevices(ctx context.Context, before time.Time) (int64, error) {
	r, err := s.db.ExecContext(ctx, `DELETE FROM devices
		WHERE online=0 AND known=0 AND label='' AND tags='' AND last_seen < ?`, before.Unix())
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// Backup writes a consistent copy of the database to path (it must not exist
// yet). The copy is created with mode 0600: it holds personal data and secrets.
func (s *Store) Backup(ctx context.Context, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	// Lstat, not Stat: a dangling symlink planted at the name would otherwise
	// look free, and the copy would be written wherever it points
	if _, err := os.Lstat(abs); err == nil {
		return fmt.Errorf("%s already exists; choose a new file name", abs)
	}
	old := syscallUmask(0o077) // the copy is never readable by others, not even for an instant
	defer syscallUmask(old)
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, abs); err != nil {
		_ = os.Remove(abs) // never leave a half-written copy that looks like a backup
		return err
	}
	// something swapped in during the copy (a symlink race in a shared folder) is refused
	if fi, err := os.Lstat(abs); err != nil || !fi.Mode().IsRegular() {
		_ = os.Remove(abs)
		return fmt.Errorf("%s is not a regular file after the backup; refusing it", abs)
	}
	return os.Chmod(abs, 0o600)
}
