//go:build linux

package tools

import (
	"context"
	"net/netip"
	"testing"
	"time"
)

// On loopback the "destination" answers with ICMP port unreachable at once,
// which exercises the whole unprivileged path: TTL, IP_RECVERR, the error queue.
func TestTracerouteReachesLoopback(t *testing.T) {
	hops, err := Traceroute(context.Background(), netip.MustParseAddr("127.0.0.1"), 5, time.Second)
	if err != nil {
		t.Skipf("this environment does not allow it: %v", err)
	}
	if len(hops) != 1 || !hops[0].Reached || hops[0].Addr != "127.0.0.1" || hops[0].TTL != 1 {
		t.Fatalf("%+v", hops)
	}
}

func TestTracerouteValidation(t *testing.T) {
	if _, err := Traceroute(context.Background(), netip.MustParseAddr("::1"), 5, time.Second); err == nil {
		t.Fatal("IPv6 must be refused")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Traceroute(ctx, netip.MustParseAddr("127.0.0.1"), 5, time.Second); err == nil {
		t.Fatal("a cancelled context must stop the trace")
	}
}
