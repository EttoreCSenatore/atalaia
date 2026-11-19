package dhcpsniff

import (
	"context"
	"net/netip"
	"path/filepath"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/store"
)

func TestLearnerNamesAnonymousDevicesAndKeepsWhatIsKnown(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	mk := func(mac string, name string, rnd bool) {
		_, _ = st.Record(ctx, time.Now(), []store.Observation{{MAC: mac, IP: netip.MustParseAddr("192.168.1.50"), Name: name, Randomized: rnd, Kind: map[bool]string{true: "mobile"}[rnd]}}, time.Hour)
	}
	mk("da:00:00:00:00:01", "", true)           // anonymous random-MAC device
	mk("da:00:00:00:00:02", "known-name", true) // already has a name
	l := &Learner{Store: st}

	l.Handle(ctx, Client{MAC: "da:00:00:00:00:01", Hostname: "Marias-iPhone", VendorClass: ""})
	d, _, _ := st.DeviceByMAC(ctx, "da:00:00:00:00:01")
	if d.Name != "Marias-iPhone" || d.Kind != "phone" {
		t.Fatalf("anonymous device not identified: %+v", d)
	}
	l.Handle(ctx, Client{MAC: "da:00:00:00:00:02", Hostname: "something-else", VendorClass: "android-dhcp-14"})
	d, _, _ = st.DeviceByMAC(ctx, "da:00:00:00:00:02")
	if d.Name != "known-name" || d.Kind != "phone" {
		t.Fatalf("a known name must be kept, the generic kind refined: %+v", d)
	}
	// the user's own label is never touched (different column) and a precise kind is not replaced by a weaker one
	_, _, _ = st.UpdateDevice(ctx, "da:00:00:00:00:01", "My phone", nil, true)
	l.Handle(ctx, Client{MAC: "da:00:00:00:00:01", Hostname: "x", VendorClass: "MSFT 5.0"})
	d, _, _ = st.DeviceByMAC(ctx, "da:00:00:00:00:01")
	if d.Label != "My phone" || d.Kind != "phone" || d.Name != "Marias-iPhone" {
		t.Fatalf("%+v", d)
	}

	// an announcement from a device not found yet is applied after it shows up
	l.Handle(ctx, Client{MAC: "da:00:00:00:00:03", Hostname: "late-comer-laptop"})
	mk("da:00:00:00:00:03", "", true)
	l.Flush(ctx)
	d, _, _ = st.DeviceByMAC(ctx, "da:00:00:00:00:03")
	if d.Name != "late-comer-laptop" || d.Kind != "computer" {
		t.Fatalf("pending announcement lost: %+v", d)
	}
	if l.Learned() < 3 {
		t.Fatalf("learned counter %d", l.Learned())
	}
}
