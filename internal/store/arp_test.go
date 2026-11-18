package store

import (
	"context"
	"fmt"
	"net/netip"
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
