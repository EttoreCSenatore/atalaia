package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Settings written by the backup routine (Unix seconds and "1"/"0").
const (
	LastRestoreTestKey   = "last_restore_test"
	LastRestoreTestOKKey = "last_restore_test_ok"
)

var backupName = regexp.MustCompile(`^atalaia-\d{8}-\d{6}\.db$`)

// BackupTo writes a new backup into dir (created 0700) named after now,
// records the time and keeps only the newest keep backups made this way:
// other files in dir are never touched. It returns the new file's path.
func (s *Store) BackupTo(ctx context.Context, dir string, keep int, now time.Time) (string, error) {
	if keep < 1 {
		return "", errors.New("keep must be at least 1")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "atalaia-"+now.UTC().Format("20060102-150405")+".db")
	if err := s.Backup(ctx, path); err != nil {
		return "", err
	}
	if err := s.SetSetting(ctx, LastBackupKey, strconv.FormatInt(now.Unix(), 10)); err != nil {
		return path, err
	}
	list, err := Backups(dir)
	if err != nil {
		return path, err
	}
	for len(list) > keep {
		if err := os.Remove(list[0]); err != nil && !errors.Is(err, os.ErrNotExist) {
			return path, err
		}
		list = list[1:]
	}
	return path, nil
}

// Backups lists the backups made by BackupTo in dir, oldest first.
func Backups(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.Type().IsRegular() && backupName.MatchString(e.Name()) {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out) // the names sort by time
	return out, nil
}

// RestoreCheck is what a restore test found in a backup.
type RestoreCheck struct {
	Schema  int // schema version of the backup before it was opened
	Devices int
}

// VerifyBackup does a real restore of path into a scratch copy inside
// workDir (removed afterwards; the backup itself is never opened for
// writing): it opens the copy as Atalaia would, which runs any pending
// migrations, then runs SQLite's integrity check and reads the inventory.
func VerifyBackup(ctx context.Context, path, workDir string) (RestoreCheck, error) {
	var rc RestoreCheck
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return rc, err
	}
	removeStaleScratch(workDir)
	tmp, err := os.MkdirTemp(workDir, ".restore-test-")
	if err != nil {
		return rc, err
	}
	defer os.RemoveAll(tmp)
	copyPath := filepath.Join(tmp, "restore.db")
	if err := copyFile(path, copyPath); err != nil {
		return rc, err
	}
	st, err := Open(copyPath)
	if err != nil {
		return rc, fmt.Errorf("the backup does not open: %w", err)
	}
	defer st.Close()
	var ic string
	if err := st.db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&ic); err != nil {
		return rc, err
	}
	if ic != "ok" {
		return rc, fmt.Errorf("integrity check: %s", ic)
	}
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&rc.Devices); err != nil {
		return rc, err
	}
	rc.Schema, err = schemaOf(path)
	return rc, err
}

// RecordRestoreTest stores the outcome of a restore test.
func (s *Store) RecordRestoreTest(ctx context.Context, at time.Time, ok bool) error {
	v := "0"
	if ok {
		v = "1"
	}
	if err := s.SetSetting(ctx, LastRestoreTestOKKey, v); err != nil {
		return err
	}
	return s.SetSetting(ctx, LastRestoreTestKey, strconv.FormatInt(at.Unix(), 10))
}

// LastRestoreTest returns when the last restore test ran and whether it passed.
func (s *Store) LastRestoreTest(ctx context.Context) (time.Time, bool, error) {
	n, err := s.SettingInt(ctx, LastRestoreTestKey, 0)
	if err != nil || n <= 0 {
		return time.Time{}, false, err
	}
	ok, err := s.Setting(ctx, LastRestoreTestOKKey, "0")
	return time.Unix(int64(n), 0), ok == "1", err
}

// schemaOf reads the schema version of a database file from its header
// (SQLite keeps PRAGMA user_version at byte offset 60, big-endian).
func schemaOf(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var h [100]byte
	if _, err := io.ReadFull(f, h[:]); err != nil {
		return 0, errors.New("not a database file")
	}
	if string(h[:16]) != "SQLite format 3\x00" {
		return 0, errors.New("not a database file")
	}
	return int(h[60])<<24 | int(h[61])<<16 | int(h[62])<<8 | int(h[63]), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// removeStaleScratch deletes scratch copies left by a restore test that was
// killed half-way (older than an hour, so a test running now is left alone).
func removeStaleScratch(dir string) {
	old, _ := filepath.Glob(filepath.Join(dir, ".restore-test-*"))
	for _, p := range old {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() && time.Since(fi.ModTime()) > time.Hour {
			_ = os.RemoveAll(p)
		}
	}
}

// preMigrationBackup is replaced in tests to simulate a failing copy.
var preMigrationBackup = (*Store).backupBeforeMigration

// PreMigrationKeep is how many pre-migration copies are kept next to the database.
const PreMigrationKeep = 3

// backupBeforeMigration copies the database next to itself before its schema
// is upgraded from version v (<name>-pre-migration-schema<v>-<time>.db, 0600),
// keeping the newest PreMigrationKeep such copies. A failed copy stops the
// upgrade: Atalaia never migrates a database it could not back up.
func (s *Store) backupBeforeMigration(v int, now time.Time) (string, error) {
	dir, file := filepath.Split(s.path)
	base := strings.TrimSuffix(file, filepath.Ext(file))
	path := filepath.Join(dir, fmt.Sprintf("%s-pre-migration-schema%d-%s.db", base, v, now.UTC().Format("20060102-150405")))
	if err := s.Backup(context.Background(), path); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(base) + `-pre-migration-schema\d+-\d{8}-\d{6}\.db$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return path, nil
	}
	var copies []string
	for _, e := range entries {
		if e.Type().IsRegular() && re.MatchString(e.Name()) {
			copies = append(copies, e.Name())
		}
	}
	sort.Slice(copies, func(i, j int) bool { return copies[i][len(copies[i])-18:] < copies[j][len(copies[j])-18:] }) // by time
	for len(copies) > PreMigrationKeep {
		_ = os.Remove(filepath.Join(dir, copies[0]))
		copies = copies[1:]
	}
	return path, nil
}
