package store

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func ip(s string) netip.Addr { return netip.MustParseAddr(s) }

func gw(mac string) Observation {
	return Observation{MAC: mac, IP: ip("192.168.1.1"), IPs: []netip.Addr{ip("192.168.1.1")}, Gateway: true}
}

func TestGatewayMACChangeIsReportedOnce(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	r, _ := s.Record(ctx, now, []Observation{gw("aa:00:00:00:00:01")}, time.Hour)
	if len(r.Findings) != 0 {
		t.Fatalf("the first sight of the gateway is not a finding: %+v", r.Findings)
	}
	r, _ = s.Record(ctx, now.Add(time.Minute), []Observation{gw("aa:00:00:00:00:01")}, time.Hour)
	if len(r.Findings) != 0 {
		t.Fatalf("unchanged: %+v", r.Findings)
	}
	r, err := s.Record(ctx, now.Add(2*time.Minute), []Observation{gw("bb:00:00:00:00:02")}, time.Hour)
	if err != nil || len(r.Findings) != 1 || r.Findings[0].Kind != EventGatewayChanged {
		t.Fatalf("want one gateway_changed, got %+v %v", r.Findings, err)
	}
	if r.Findings[0].Device.MAC != "bb:00:00:00:00:02" || r.Findings[0].Detail != "192.168.1.1: aa:00:00:00:00:01 -> bb:00:00:00:00:02" {
		t.Fatalf("%+v", r.Findings[0])
	}
	r, _ = s.Record(ctx, now.Add(3*time.Minute), []Observation{gw("bb:00:00:00:00:02")}, time.Hour)
	if len(r.Findings) != 0 {
		t.Fatalf("reported again: %+v", r.Findings)
	}
	evs, _ := s.Events(ctx, 10)
	if evs[0].Kind != EventGatewayChanged {
		t.Fatalf("event not stored: %+v", evs[0])
	}
}

func multi(mac string, n int) Observation {
	o := Observation{MAC: mac, IP: ip("192.168.1.10")}
	for i := 0; i < n; i++ {
		o.IPs = append(o.IPs, ip(fmt.Sprintf("192.168.1.%d", 10+i)))
	}
	return o
}

func TestMultiIPFlaggedWhenItStartsNotEveryScan(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	m := "cc:00:00:00:00:03"
	r, _ := s.Record(ctx, now, []Observation{multi(m, 2)}, time.Hour)
	if len(r.Findings) != 0 {
		t.Fatalf("two addresses are fine: %+v", r.Findings)
	}
	r, _ = s.Record(ctx, now.Add(time.Minute), []Observation{multi(m, 3)}, time.Hour)
	if len(r.Findings) != 1 || r.Findings[0].Kind != EventMultiIP {
		t.Fatalf("want multi_ip: %+v", r.Findings)
	}
	r, _ = s.Record(ctx, now.Add(2*time.Minute), []Observation{multi(m, 3)}, time.Hour)
	if len(r.Findings) != 0 {
		t.Fatalf("repeated: %+v", r.Findings)
	}
	// it stops, then starts again: reported again
	_, _ = s.Record(ctx, now.Add(3*time.Minute), []Observation{multi(m, 1)}, time.Hour)
	r, _ = s.Record(ctx, now.Add(4*time.Minute), []Observation{multi(m, 4)}, time.Hour)
	if len(r.Findings) != 1 {
		t.Fatalf("should be reported again: %+v", r.Findings)
	}
}

func TestKnownDevicesMayHaveManyAddresses(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	m := "cc:00:00:00:00:04"
	_, _ = s.Record(ctx, now, []Observation{multi(m, 1)}, time.Hour)
	if _, ok, err := s.UpdateDevice(ctx, m, "", nil, true); err != nil || !ok {
		t.Fatal(err)
	}
	r, _ := s.Record(ctx, now.Add(time.Minute), []Observation{multi(m, 5)}, time.Hour)
	if len(r.Findings) != 0 {
		t.Fatalf("known device flagged: %+v", r.Findings)
	}
}

func TestMigrationKeepsEventsAndAllowsNewKinds(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	_, _ = s.Record(ctx, now, []Observation{{MAC: "dd:00:00:00:00:05", IP: ip("192.168.1.9")}}, time.Hour)
	evs, err := s.Events(ctx, 5)
	if err != nil || len(evs) != 1 || evs[0].Kind != EventJoined {
		t.Fatalf("%+v %v", evs, err)
	}
}

func TestAuditAndPortResults(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	mac := "ee:00:00:00:00:06"
	_, _ = s.Record(ctx, now, []Observation{{MAC: mac, IP: ip("192.168.1.9")}}, time.Hour)

	if _, ok, _ := s.Ports(ctx, mac); ok {
		t.Fatal("never scanned")
	}
	if err := s.SavePorts(ctx, mac, now, []OpenPort{{22, "ssh"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SavePorts(ctx, mac, now.Add(time.Hour), nil); err != nil { // overwrite with none open
		t.Fatal(err)
	}
	r, ok, err := s.Ports(ctx, mac)
	if err != nil || !ok || len(r.Ports) != 0 || !r.At.Equal(now.Add(time.Hour)) {
		t.Fatalf("%+v %v %v", r, ok, err)
	}
	_ = s.SavePorts(ctx, "00:00:00:00:00:00", now, nil) // unknown device: no-op
	if _, ok, _ := s.Ports(ctx, "00:00:00:00:00:00"); ok {
		t.Fatal("saved for an unknown device")
	}

	_ = s.Audit(ctx, now, "portscan", mac, "2 open")
	_ = s.Audit(ctx, now.Add(time.Minute), "wake", mac, "")
	a, _ := s.Audits(ctx, 10)
	if len(a) != 2 || a[0].Action != "wake" {
		t.Fatalf("%+v", a)
	}
	if n, _ := s.PurgeEvents(ctx, now.Add(10*time.Minute)); n < 0 {
		t.Fatal(n)
	}
	if a, _ := s.Audits(ctx, 10); len(a) != 0 {
		t.Fatalf("audit not purged: %+v", a)
	}
}

func TestDHCPBaselineThenNewServerIsFlaggedOnce(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	// the future rogue server is also a device on the network
	_, _ = s.Record(ctx, now, []Observation{{MAC: "ff:00:00:00:00:07", IP: ip("192.168.1.99")}}, time.Hour)

	r, err := s.RecordDHCP(ctx, now, []DHCPObservation{{IP: "192.168.1.1", Detail: "{}"}})
	if err != nil || !r.Baseline || len(r.New) != 0 {
		t.Fatalf("baseline: %+v %v", r, err)
	}
	r, _ = s.RecordDHCP(ctx, now.Add(time.Minute), []DHCPObservation{{IP: "192.168.1.1"}})
	if r.Baseline || len(r.New) != 0 {
		t.Fatalf("same server: %+v", r)
	}
	r, _ = s.RecordDHCP(ctx, now.Add(2*time.Minute), []DHCPObservation{{IP: "192.168.1.1"}, {IP: "192.168.1.99", Detail: "{}"}})
	if len(r.New) != 1 || r.New[0].IP != "192.168.1.99" {
		t.Fatalf("new server: %+v", r)
	}
	r, _ = s.RecordDHCP(ctx, now.Add(3*time.Minute), []DHCPObservation{{IP: "192.168.1.99"}})
	if len(r.New) != 0 {
		t.Fatalf("flagged twice: %+v", r)
	}
	evs, _ := s.Events(ctx, 5)
	if evs[0].Kind != EventRogueDHCP || evs[0].MAC != "ff:00:00:00:00:07" {
		t.Fatalf("event: %+v", evs[0])
	}
	if list, _ := s.DHCPServers(ctx); len(list) != 2 {
		t.Fatalf("%+v", list)
	}
	// an empty probe never creates a baseline; forgetting resets it
	_ = s.ForgetDHCPServers(ctx)
	if r, _ := s.RecordDHCP(ctx, now, nil); r.Baseline {
		t.Fatal("an empty probe is not a baseline")
	}
	if r, _ := s.RecordDHCP(ctx, now, []DHCPObservation{{IP: "10.0.0.1"}}); !r.Baseline {
		t.Fatal("after forgetting, the next probe is a baseline")
	}
}

func TestWhatADeviceSaysAboutItselfIsKeptAndOnlyReplacedByNewInformation(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	mac := "aa:11:22:00:00:09"
	_, _ = s.Record(ctx, now, []Observation{{MAC: mac, IP: ip("192.168.1.30"), Model: "Chromecast", Kind: "tv", Services: []string{"_googlecast._tcp", "_airplay._tcp"}}}, time.Hour)
	// a scan where the device stayed silent must not erase it
	_, _ = s.Record(ctx, now.Add(time.Minute), []Observation{{MAC: mac, IP: ip("192.168.1.30")}}, time.Hour)
	d, _, _ := getByMAC(ctx, s.db, mac)
	if d.Model != "Chromecast" || d.Kind != "tv" || len(d.Services) != 2 {
		t.Fatalf("erased by silence: %+v", d)
	}
	_, _ = s.Record(ctx, now.Add(2*time.Minute), []Observation{{MAC: mac, IP: ip("192.168.1.30"), Kind: "speaker", Services: []string{"_spotify-connect._tcp"}}}, time.Hour)
	d, _, _ = getByMAC(ctx, s.db, mac)
	if d.Kind != "speaker" || d.Model != "Chromecast" || len(d.Services) != 1 {
		t.Fatalf("new information not applied: %+v", d)
	}
}

func TestDeletePruneAndBackup(t *testing.T) {
	s, ctx, now := open(t), context.Background(), time.Unix(1700000000, 0)
	rec := func(mac, ipa string, at time.Time) {
		_, _ = s.Record(ctx, at, []Observation{{MAC: mac, IP: ip(ipa)}}, time.Hour)
	}
	rec("aa:00:00:00:00:01", "192.168.1.10", now.AddDate(0, 0, -40)) // old, untouched: pruned
	rec("aa:00:00:00:00:02", "192.168.1.11", now.AddDate(0, 0, -40)) // old but known: kept
	rec("aa:00:00:00:00:03", "192.168.1.12", now)                    // recent: kept
	_, _, _ = s.UpdateDevice(ctx, "aa:00:00:00:00:02", "", nil, true)
	_, _ = s.Record(ctx, now, []Observation{{MAC: "aa:00:00:00:00:03", IP: ip("192.168.1.12")}}, time.Hour) // the others go offline
	if n, err := s.PruneDevices(ctx, now.AddDate(0, 0, -30)); err != nil || n != 1 {
		t.Fatalf("pruned %d %v", n, err)
	}
	if _, ok, _ := s.DeviceByMAC(ctx, "aa:00:00:00:00:01"); ok {
		t.Fatal("old untouched device kept")
	}
	if ok, err := s.DeleteDevice(ctx, "aa:00:00:00:00:03"); !ok || err != nil {
		t.Fatalf("delete: %v %v", ok, err)
	}
	if ok, _ := s.DeleteDevice(ctx, "aa:00:00:00:00:03"); ok {
		t.Fatal("deleted twice")
	}
	evs, _ := s.Events(ctx, 50)
	for _, e := range evs {
		if e.MAC == "aa:00:00:00:00:03" || e.MAC == "aa:00:00:00:00:01" {
			t.Fatalf("history of a forgotten device kept: %+v", e)
		}
	}

	dst := filepath.Join(t.TempDir(), "backup.db")
	if err := s.Backup(ctx, dst); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(dst)
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("backup mode: %v %v", st, err)
	}
	if err := s.Backup(ctx, dst); err == nil {
		t.Fatal("a backup must not overwrite an existing file")
	}
	b, err := Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if list, _ := b.Devices(ctx); len(list) != 1 || list[0].MAC != "aa:00:00:00:00:02" {
		t.Fatalf("backup content: %+v", list)
	}
}
