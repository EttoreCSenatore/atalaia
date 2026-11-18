package alerts

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/EttoreCSenatore/atalaia/internal/notify"
	"github.com/EttoreCSenatore/atalaia/internal/store"
)

type fake struct {
	mu   sync.Mutex
	sent []string
	fail bool
}

func (f *fake) Send(_ context.Context, c notify.Channel, m notify.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, c.Name+"|"+m.Title+"|"+m.Body)
	if f.fail {
		return context.DeadlineExceeded
	}
	return nil
}

func newService(t *testing.T) (*Service, *fake) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	f := &fake{}
	return &Service{Store: st, Sender: f, Logf: func(string, ...any) {}}, f
}

var dev = store.Device{MAC: "aa:bb:cc:00:00:01", IP: "192.168.1.9", Vendor: "Acme", Name: "cam.local"}

func TestComposeRules(t *testing.T) {
	res := store.Result{New: []store.Device{dev}, WentOffline: []store.Device{dev, {MAC: "x", Known: true, Label: "Printer"}}}

	if got := Compose(res, Rules{NewDevice: true}, false); len(got) != 1 || !strings.Contains(got[0].Body, "cam.local") {
		t.Fatalf("new device: %+v", got)
	}
	if got := Compose(res, Rules{NewDevice: true}, true); len(got) != 0 {
		t.Fatalf("the first scan is the baseline, not news: %+v", got)
	}
	got := Compose(res, Rules{KnownOffline: true}, false)
	if len(got) != 1 || !strings.Contains(got[0].Body, "Printer") || strings.Contains(got[0].Body, "cam.local") {
		t.Fatalf("only known devices are reported offline: %+v", got)
	}
	if got := Compose(res, Rules{}, false); len(got) != 0 {
		t.Fatalf("nothing enabled: %+v", got)
	}
}

func TestComposeManyDevicesIsBounded(t *testing.T) {
	var many []store.Device
	for i := 0; i < 40; i++ {
		many = append(many, dev)
	}
	m := Compose(store.Result{New: many}, Rules{NewDevice: true}, false)
	if n := strings.Count(m[0].Body, "\n"); n > 11 {
		t.Fatalf("too long: %d lines", n)
	}
	if !strings.Contains(m[0].Title, "40") {
		t.Fatalf("title: %s", m[0].Title)
	}
}

func TestHandleSendsToEnabledChannelsOnly(t *testing.T) {
	s, f := newService(t)
	ctx := context.Background()
	err := s.SaveChannels(ctx, []notify.Channel{
		{Type: notify.TypeWebhook, Name: "on", Enabled: true, URL: "http://x.test/h"},
		{Type: notify.TypeWebhook, Name: "off", Enabled: false, URL: "http://x.test/h"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s.Handle(ctx, store.Result{New: []store.Device{dev}}, false) // default rules: new device on
	if len(f.sent) != 1 || !strings.HasPrefix(f.sent[0], "on|") {
		t.Fatalf("sent: %v", f.sent)
	}
	// a failing channel does not panic or stop the others
	f.fail = true
	s.Handle(ctx, store.Result{New: []store.Device{dev}}, false)
}

func TestRulesDefaultsAndRoundTrip(t *testing.T) {
	s, _ := newService(t)
	ctx := context.Background()
	r, _ := s.Rules(ctx)
	if !r.NewDevice || r.KnownOffline {
		t.Fatalf("defaults: %+v", r)
	}
	_ = s.SaveRules(ctx, Rules{NewDevice: false, KnownOffline: true})
	if r, _ := s.Rules(ctx); r.NewDevice || !r.KnownOffline {
		t.Fatalf("saved: %+v", r)
	}
}

func TestSaveChannelsLimits(t *testing.T) {
	s, _ := newService(t)
	ctx := context.Background()
	var many []notify.Channel
	for i := 0; i < MaxChannels+1; i++ {
		many = append(many, notify.Channel{Type: notify.TypeWebhook, URL: "http://x.test"})
	}
	if err := s.SaveChannels(ctx, many); err == nil {
		t.Fatal("expected the limit to apply")
	}
	dup := []notify.Channel{{ID: "a", Type: notify.TypeWebhook, URL: "http://x.test"}, {ID: "a", Type: notify.TypeWebhook, URL: "http://x.test"}}
	if err := s.SaveChannels(ctx, dup); err == nil {
		t.Fatal("expected duplicate ids to be refused")
	}
}
