package store

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/releases"
)

// Released migrations must never change: a database already upgraded by them
// would not get the new text. testdata/migrations.sha256 has one hash per
// migration; adding a migration means adding its line (the failure says which).
func TestMigrationsAreAppendOnly(t *testing.T) {
	f, err := os.Open("testdata/migrations.sha256")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var want []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" && !strings.HasPrefix(line, "#") {
			want = append(want, strings.Fields(line)[0])
		}
	}
	for i, m := range migrations {
		sum := sha256.Sum256([]byte(m))
		got := hex.EncodeToString(sum[:])
		if i >= len(want) {
			t.Errorf("migration %d is new: add this line to testdata/migrations.sha256:\n%s  migration-%d", i+1, got, i+1)
			continue
		}
		if got != want[i] {
			t.Errorf("migration %d was changed after it was released: add a new migration instead", i+1)
		}
	}
	if len(want) > len(migrations) {
		t.Errorf("testdata/migrations.sha256 lists %d migrations, the code has %d: a migration was removed", len(want), len(migrations))
	}
}

type snapshot struct {
	schema  int
	devices []string // mac|ip|name|online
	events  int
}

func readSnapshot(t *testing.T, path string) snapshot {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var s snapshot
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&s.schema); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT mac, ip, name, online FROM devices ORDER BY mac`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var mac, ip, name string
		var online int
		if err := rows.Scan(&mac, &ip, &name, &online); err != nil {
			t.Fatal(err)
		}
		s.devices = append(s.devices, fmt.Sprintf("%s|%s|%s|%d", mac, ip, name, online))
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&s.events); err != nil {
		t.Fatal(err)
	}
	return s
}

// Every published version's database (made by that version's own binary,
// see scripts/gerar-bancos-de-versoes.py) opens with this code: it is backed
// up first, migrated to the latest schema, passes the integrity check, keeps
// its devices and events, and takes new scans.
func TestUpgradeFromReleasedDatabases(t *testing.T) {
	fixtures, _ := filepath.Glob("testdata/released/v*.db")
	if len(fixtures) == 0 {
		t.Fatal("no released databases in testdata/released")
	}
	have := map[string]bool{}
	for _, f := range fixtures {
		have[strings.TrimSuffix(filepath.Base(f), ".db")] = true
	}
	rs, err := releases.All()
	if err != nil {
		t.Fatal(err)
	}
	// the newest release may not have its database yet (it is made from the published binary)
	for _, r := range rs[1:] {
		if !have["v"+r.Version] {
			t.Errorf("no database for v%s: run python3 -I scripts/gerar-bancos-de-versoes.py v%s", r.Version, r.Version)
		}
	}
	ctx := context.Background()
	for _, fx := range fixtures {
		t.Run(filepath.Base(fx), func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "atalaia.db")
			if err := copyFile(fx, path); err != nil {
				t.Fatal(err)
			}
			before := readSnapshot(t, path)
			if len(before.devices) == 0 || before.events == 0 {
				t.Fatalf("fixture has no data: %+v", before)
			}
			s, err := Open(path)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer s.Close()
			var ic string
			if err := s.db.QueryRow(`PRAGMA integrity_check`).Scan(&ic); err != nil || ic != "ok" {
				t.Fatalf("integrity: %q %v", ic, err)
			}
			after := readSnapshot(t, path)
			if after.schema != len(migrations) {
				t.Fatalf("schema %d after upgrade, want %d", after.schema, len(migrations))
			}
			if strings.Join(after.devices, "\n") != strings.Join(before.devices, "\n") || after.events != before.events {
				t.Fatalf("data changed:\nbefore %+v\nafter  %+v", before, after)
			}
			pre, _ := filepath.Glob(filepath.Join(dir, "atalaia-pre-migration-schema*.db"))
			if before.schema < len(migrations) {
				if len(pre) != 1 || readSnapshot(t, pre[0]).schema != before.schema {
					t.Fatalf("pre-migration backup: %v", pre)
				}
				if fi, _ := os.Stat(pre[0]); fi.Mode().Perm() != 0o600 {
					t.Fatalf("backup mode %v", fi.Mode().Perm())
				}
			} else if len(pre) != 0 {
				t.Fatalf("backup made without a migration: %v", pre)
			}
			// the upgraded database works: a new scan, settings, the app's queries
			if _, err := s.Record(ctx, time.Now(), []Observation{{MAC: "aa:00:00:00:00:09", IP: netip.MustParseAddr("10.123.45.9")}}, time.Hour); err != nil {
				t.Fatal(err)
			}
			if devs, err := s.Devices(ctx); err != nil || len(devs) != len(before.devices)+1 {
				t.Fatalf("devices after scan: %d %v", len(devs), err)
			}
			if err := s.SetSetting(ctx, "retention_days", "30"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPreMigrationBackupsAreRotated(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "atalaia.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	t0 := time.Date(2026, 11, 22, 3, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		if _, err := s.backupBeforeMigration(4+i%2, t0.Add(time.Duration(i)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := filepath.Glob(filepath.Join(dir, "atalaia-pre-migration-*.db"))
	sort.Strings(got)
	if len(got) != PreMigrationKeep {
		t.Fatalf("kept %v", got)
	}
	for _, g := range got {
		if strings.HasSuffix(g, "030000.db") || strings.HasSuffix(g, "030100.db") {
			t.Fatalf("an old copy survived: %v", got)
		}
	}
}

func TestNoUpgradeWithoutBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "atalaia.db")
	if err := copyFile("testdata/released/v0.1.0.db", path); err != nil {
		t.Fatal(err)
	}
	orig := preMigrationBackup
	preMigrationBackup = func(*Store, int, time.Time) (string, error) { return "", fmt.Errorf("disk full") }
	defer func() { preMigrationBackup = orig }()
	if _, err := Open(path); err == nil || !strings.Contains(err.Error(), "before upgrading") {
		t.Fatalf("upgraded a database it could not back up: %v", err)
	}
	if got := readSnapshot(t, path).schema; got != 3 {
		t.Fatalf("schema changed to %d without a backup", got)
	}
}
