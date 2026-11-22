package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupToRotatesAndVerify(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Record(ctx, time.Now(), []Observation{obs("aa:00:00:00:00:01", "192.168.1.10", "pi")}, time.Hour); err != nil {
		t.Fatal(err)
	}
	bk := filepath.Join(dir, "backups")
	other := filepath.Join(bk, "mine.db")
	if err := os.MkdirAll(bk, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 11, 20, 3, 0, 0, 0, time.UTC)
	var last string
	for i := 0; i < 5; i++ {
		if last, err = s.BackupTo(ctx, bk, 3, start.Add(time.Duration(i)*24*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := Backups(bk)
	if len(list) != 3 || list[2] != last || !strings.HasSuffix(list[0], "atalaia-20261122-030000.db") {
		t.Fatalf("rotation kept %v", list)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatal("rotation removed a file it did not make")
	}
	if at, _ := s.LastBackup(ctx); !at.Equal(start.Add(4 * 24 * time.Hour)) {
		t.Fatalf("last backup = %v", at)
	}
	if fi, _ := os.Stat(bk); fi.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode %v", fi.Mode().Perm())
	}

	stale := filepath.Join(bk, ".restore-test-old")
	if err := os.MkdirAll(stale, 0o700); err != nil {
		t.Fatal(err)
	}
	long := time.Now().Add(-2 * time.Hour)
	_ = os.Chtimes(stale, long, long)
	fresh := filepath.Join(bk, ".restore-test-running")
	if err := os.MkdirAll(fresh, 0o700); err != nil {
		t.Fatal(err)
	}
	RemoveStaleScratch(bk)
	if _, err := os.Stat(stale); err == nil {
		t.Fatal("stale scratch copy kept")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatal("a running restore test's scratch copy was removed")
	}
	_ = os.Remove(fresh)
	rc, err := VerifyBackup(ctx, last, bk)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Devices != 1 || rc.Schema != len(migrations) {
		t.Fatalf("check = %+v", rc)
	}
	if left, _ := filepath.Glob(filepath.Join(bk, ".restore-test-*")); len(left) != 0 {
		t.Fatalf("scratch copy left behind: %v", left)
	}
	// the backup itself was not changed by the test
	before, _ := os.ReadFile(last)
	if _, err := VerifyBackup(ctx, last, bk); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(last)
	if string(before) != string(after) {
		t.Fatal("restore test modified the backup")
	}

	// a damaged backup fails
	bad := filepath.Join(dir, "bad.db")
	b := append([]byte(nil), before...)
	for i := 4096; i < len(b) && i < 4096+2048; i++ {
		b[i] ^= 0xff
	}
	if err := os.WriteFile(bad, b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyBackup(ctx, bad, bk); err == nil {
		t.Fatal("damaged backup passed")
	}
	if err := os.WriteFile(bad, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyBackup(ctx, bad, bk); err == nil {
		t.Fatal("non-database passed")
	}

	if err := s.RecordRestoreTest(ctx, start, true); err != nil {
		t.Fatal(err)
	}
	if at, ok, err := s.LastRestoreTest(ctx); err != nil || !ok || !at.Equal(start) {
		t.Fatalf("restore test = %v %v %v", at, ok, err)
	}
	if _, err := s.BackupTo(ctx, bk, 0, start); err == nil {
		t.Fatal("keep 0 accepted")
	}
}

func TestBackupRefusesSymlinksAndSharedFolders(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	// a dangling symlink at the name must not be followed
	outside := filepath.Join(t.TempDir(), "planted.db")
	link := filepath.Join(dir, "copy.db")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if err := s.Backup(ctx, link); err == nil {
		t.Fatal("backup through a symlink accepted")
	}
	if _, err := os.Stat(outside); err == nil {
		t.Fatal("the copy was written through the symlink")
	}
	// a folder other users can write to is refused
	shared := filepath.Join(dir, "shared")
	if err := os.Mkdir(shared, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(shared, 0o777); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BackupTo(ctx, shared, 3, time.Now()); err == nil || !strings.Contains(err.Error(), "other users") {
		t.Fatalf("shared folder accepted: %v", err)
	}
	if list, _ := Backups(shared); len(list) != 0 {
		t.Fatalf("a backup was written: %v", list)
	}
}

func TestUntrustedSchemaIsOff(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var v int
	if err := s.db.QueryRow(`PRAGMA trusted_schema`).Scan(&v); err != nil || v != 0 {
		t.Fatalf("trusted_schema = %d, %v", v, err)
	}
}
