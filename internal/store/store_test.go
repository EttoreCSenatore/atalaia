package store

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "sub", "atalaia.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func obs(mac, ip, name string) Observation {
	return Observation{MAC: mac, IP: netip.MustParseAddr(ip), Name: name, Vendor: "Acme"}
}

var t0 = time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

func TestFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "atalaia.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for p, want := range map[string]os.FileMode{path: 0o600, filepath.Dir(path): 0o700} {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != want {
			t.Errorf("%s mode = %v, want %v", p, st.Mode().Perm(), want)
		}
	}
}

func TestJoinOnlineOfflineLifecycle(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	grace := 10 * time.Minute

	// 1. first scan: two new devices
	r, err := s.Record(ctx, t0, []Observation{obs("a8:00:00:00:00:01", "192.168.1.10", "tv"), obs("a8:00:00:00:00:02", "192.168.1.11", "")}, grace)
	if err != nil || len(r.New) != 2 {
		t.Fatalf("scan 1: %+v %v", r, err)
	}

	// 2. five minutes later device 2 is missing: still online (grace period)
	r, _ = s.Record(ctx, t0.Add(5*time.Minute), []Observation{obs("a8:00:00:00:00:01", "192.168.1.10", "")}, grace)
	if len(r.WentOffline) != 0 || len(r.New) != 0 {
		t.Fatalf("scan 2 must not flap: %+v", r)
	}

	// 3. twenty minutes later it is still missing: offline
	r, _ = s.Record(ctx, t0.Add(20*time.Minute), []Observation{obs("a8:00:00:00:00:01", "192.168.1.10", "")}, grace)
	if len(r.WentOffline) != 1 || r.WentOffline[0].MAC != "a8:00:00:00:00:02" {
		t.Fatalf("scan 3: %+v", r)
	}
	// and it is not reported offline again
	r, _ = s.Record(ctx, t0.Add(30*time.Minute), []Observation{obs("a8:00:00:00:00:01", "192.168.1.10", "")}, grace)
	if len(r.WentOffline) != 0 {
		t.Fatalf("offline reported twice: %+v", r)
	}

	// 4. it comes back with a new address
	r, _ = s.Record(ctx, t0.Add(40*time.Minute), []Observation{obs("a8:00:00:00:00:01", "192.168.1.10", ""), obs("a8:00:00:00:00:02", "192.168.1.50", "")}, grace)
	if len(r.CameOnline) != 1 || len(r.IPChanged) != 1 {
		t.Fatalf("scan 4: %+v", r)
	}

	devs, _ := s.Devices(ctx)
	if len(devs) != 2 || !devs[0].Online || !devs[1].Online {
		t.Fatalf("devices: %+v", devs)
	}
	if devs[0].Name != "tv" {
		t.Fatalf("an empty name must not erase a known one: %+v", devs[0])
	}
	if devs[1].IP != "192.168.1.50" {
		t.Fatalf("ip not updated: %+v", devs[1])
	}

	evs, _ := s.Events(ctx, 50)
	kinds := map[string]int{}
	for _, e := range evs {
		kinds[e.Kind]++
	}
	if kinds[EventJoined] != 2 || kinds[EventOffline] != 1 || kinds[EventOnline] != 1 || kinds[EventIPChanged] != 1 {
		t.Fatalf("events: %v", kinds)
	}
	if evs[0].At.Before(evs[len(evs)-1].At) {
		t.Fatal("events must be newest first")
	}
}

func TestReopenKeepsDataAndRejectsNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "atalaia.db")
	s, _ := Open(path)
	_, _ = s.Record(context.Background(), t0, []Observation{obs("a8:00:00:00:00:01", "192.168.1.10", "tv")}, time.Hour)
	_ = s.Close()

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if d, _ := s.Devices(context.Background()); len(d) != 1 {
		t.Fatalf("data lost: %+v", d)
	}
	_, _ = s.db.Exec(`PRAGMA user_version = 99`)
	_ = s.Close()
	if _, err := Open(path); err == nil {
		t.Fatal("must refuse a database from a newer version")
	}
}
