package dhcpwatch

import (
	"context"
	"errors"
	"net/netip"
	"path/filepath"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/dhcpprobe"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

func offer(ip string) dhcpprobe.Offer {
	a := netip.MustParseAddr(ip)
	return dhcpprobe.Offer{Server: a, From: a, Router: a}
}

func TestCheckBaselineThenNewServer(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	answers := [][]dhcpprobe.Offer{{offer("192.168.1.1")}, {offer("192.168.1.1")}, {offer("192.168.1.1"), offer("192.168.1.77")}}
	var probeErr error
	i := 0
	w := &Watcher{Store: st, Now: time.Now, Probe: func(context.Context) ([]dhcpprobe.Offer, error) {
		if probeErr != nil {
			return nil, probeErr
		}
		a := answers[i]
		i++
		return a, nil
	}}
	r, err := w.Check(context.Background())
	if err != nil || !r.Baseline || len(r.New) != 0 {
		t.Fatalf("baseline: %+v %v", r, err)
	}
	if r, _ = w.Check(context.Background()); r.Baseline || len(r.New) != 0 {
		t.Fatalf("unchanged: %+v", r)
	}
	if r, _ = w.Check(context.Background()); len(r.New) != 1 || r.New[0] != "192.168.1.77" || len(r.Offers) != 2 {
		t.Fatalf("rogue: %+v", r)
	}
	if w.Last() == nil || len(w.Last().New) != 1 {
		t.Fatal("last report not kept")
	}
	probeErr = errors.New("needs privilege")
	if r, err := w.Check(context.Background()); err == nil || r.Error == "" || w.Last().Error == "" {
		t.Fatalf("a failing probe must be reported: %+v %v", r, err)
	}
}

func TestEnabledDefaultsOffAndOverlapIsRefused(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "a.db"))
	defer st.Close()
	block := make(chan struct{})
	w := &Watcher{Store: st, Probe: func(context.Context) ([]dhcpprobe.Offer, error) { <-block; return nil, nil }}
	if w.Enabled(context.Background()) {
		t.Fatal("the scheduled check must be off by default (it is an active probe)")
	}
	_ = st.SetSetting(context.Background(), EnabledKey, "1")
	if !w.Enabled(context.Background()) {
		t.Fatal("setting not honoured")
	}
	done := make(chan struct{})
	go func() { _, _ = w.Check(context.Background()); close(done) }()
	time.Sleep(50 * time.Millisecond)
	if _, err := w.Check(context.Background()); !errors.Is(err, ErrBusy) {
		t.Fatalf("overlap: %v", err)
	}
	close(block)
	<-done
}
