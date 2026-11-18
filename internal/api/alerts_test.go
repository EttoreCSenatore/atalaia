package api

import (
	"context"
	"net/netip"

	"encoding/json"
	"github.com/EttoreCSenatore/atalaia/internal/store"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func (f *fixture) seedDevice() {
	f.t.Helper()
	obs := []store.Observation{{MAC: "b8:27:eb:00:00:01", IP: netip.MustParseAddr("192.168.1.10"), Vendor: "Raspberry Pi", Name: "pi.local"}}
	if _, err := f.st.Record(context.Background(), time.Now(), obs, time.Hour); err != nil {
		f.t.Fatal(err)
	}
}

func TestEditDevice(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	f.seedDevice()
	mac := "b8:27:eb:00:00:01"

	w := f.do("PUT", "/api/devices/"+mac, map[string]any{"label": "Living room Pi", "tags": []string{" IoT ", "iot", "Media"}, "known": true}, c)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var d deviceDTO
	_ = json.Unmarshal(w.Body.Bytes(), &d)
	if d.Label != "Living room Pi" || !d.Known || len(d.Tags) != 2 || d.Tags[0] != "iot" || d.Tags[1] != "media" {
		t.Fatalf("unexpected device: %+v", d)
	}
	// persisted and listed
	w = f.do("GET", "/api/devices", nil, c)
	var list []deviceDTO
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Label != "Living room Pi" || !list[0].Known {
		t.Fatalf("not persisted: %+v", list)
	}

	for name, tc := range map[string]struct {
		path string
		body map[string]any
		want int
	}{
		"bad mac":      {"/api/devices/nope", map[string]any{"label": "x"}, 400},
		"unknown":      {"/api/devices/aa:bb:cc:dd:ee:ff", map[string]any{"label": "x"}, 404},
		"long label":   {"/api/devices/" + mac, map[string]any{"label": strings.Repeat("a", 65)}, 400},
		"control char": {"/api/devices/" + mac, map[string]any{"label": "a\nb"}, 400},
		"comma tag":    {"/api/devices/" + mac, map[string]any{"tags": []string{"a,b"}}, 400},
		"many tags":    {"/api/devices/" + mac, map[string]any{"tags": []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}}, 400},
		"extra field":  {"/api/devices/" + mac, map[string]any{"ip": "1.1.1.1"}, 400},
	} {
		if w := f.do("PUT", tc.path, tc.body, c); w.Code != tc.want {
			t.Errorf("%s: got %d want %d (%s)", name, w.Code, tc.want, w.Body)
		}
	}
	if w := f.do("PUT", "/api/devices/"+mac, map[string]any{"label": "x"}, nil); w.Code != 401 {
		t.Errorf("anonymous edit: %d", w.Code)
	}
}

func TestAlertChannelsKeepSecretsServerSide(t *testing.T) {
	f := newFixture(t)
	c := f.setup()

	var got []string
	hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = append(got, string(b))
	}))
	defer hook.Close()

	body := map[string]any{
		"rules": map[string]bool{"new_device": true, "known_offline": true},
		"channels": []map[string]any{
			{"type": "telegram", "name": "tg", "enabled": true, "token": "123:SECRET", "chat_id": "42"},
			{"type": "webhook", "name": "hook", "enabled": true, "url": hook.URL},
		},
	}
	w := f.do("PUT", "/api/alerts", body, c)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if strings.Contains(w.Body.String(), "SECRET") {
		t.Fatal("the response leaked a secret")
	}
	var v alertsDTO
	_ = json.Unmarshal(w.Body.Bytes(), &v)
	if len(v.Channels) != 2 || !v.Channels[0].HasToken || v.Channels[0].ID == "" || !v.Rules.KnownOffline {
		t.Fatalf("unexpected view: %+v", v)
	}
	if w := f.do("GET", "/api/alerts", nil, c); strings.Contains(w.Body.String(), "SECRET") {
		t.Fatal("GET leaked a secret")
	}

	// saving again without the secret keeps the stored one
	v.Channels[0].Name = "renamed"
	again := map[string]any{"rules": v.Rules, "channels": v.Channels}
	if w := f.do("PUT", "/api/alerts", again, c); w.Code != 200 {
		t.Fatalf("resave: %d %s", w.Code, w.Body)
	}
	chans, _ := f.srv.Alerts.Channels(t.Context())
	if chans[0].Token != "123:SECRET" || chans[0].Name != "renamed" {
		t.Fatalf("secret lost: %+v", chans[0])
	}

	// test button
	hookID := chans[1].ID
	if w := f.do("POST", "/api/alerts/test", map[string]string{"id": hookID}, c); w.Code != 200 {
		t.Fatalf("test: %d %s", w.Code, w.Body)
	}
	if len(got) != 1 || !strings.Contains(got[0], "atalaia") {
		t.Fatalf("webhook got %v", got)
	}
	if w := f.do("POST", "/api/alerts/test", map[string]string{"id": "nope"}, c); w.Code != 404 {
		t.Errorf("unknown channel: %d", w.Code)
	}
}

func TestAlertChannelValidation(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	for name, ch := range map[string]map[string]any{
		"unknown type": {"type": "sms", "enabled": true},
		"bad url":      {"type": "ntfy", "url": "ftp://x/y"},
		"javascript":   {"type": "webhook", "url": "javascript:alert(1)"},
		"telegram":     {"type": "telegram", "token": "x/../y", "chat_id": "1"},
		"email host":   {"type": "email", "host": "smtp", "from": "a@b.c", "to": "d@e.f"},
		"email from":   {"type": "email", "host": "smtp:587", "from": "nope", "to": "d@e.f"},
	} {
		w := f.do("PUT", "/api/alerts", map[string]any{"rules": map[string]bool{}, "channels": []any{ch}}, c)
		if w.Code != 400 {
			t.Errorf("%s: got %d (%s)", name, w.Code, w.Body)
		}
	}
	if w := f.do("PUT", "/api/alerts", map[string]any{"rules": map[string]bool{}, "channels": []any{}}, nil); w.Code != 401 {
		t.Errorf("anonymous: %d", w.Code)
	}
}

func TestDevicesExposeWhatTheyAnnounced(t *testing.T) {
	f := newFixture(t)
	c := f.setup()
	obs := []store.Observation{{MAC: "da:00:00:00:00:01", IP: netip.MustParseAddr("192.168.1.40"), Model: "Chromecast", Kind: "tv", Services: []string{"_googlecast._tcp"}}}
	if _, err := f.st.Record(context.Background(), time.Now(), obs, time.Hour); err != nil {
		t.Fatal(err)
	}
	var list []deviceDTO
	_ = json.Unmarshal(f.do("GET", "/api/devices", nil, c).Body.Bytes(), &list)
	if len(list) != 1 || list[0].Kind != "tv" || list[0].Model != "Chromecast" || len(list[0].Services) != 1 {
		t.Fatalf("%+v", list)
	}
	raw := f.do("GET", "/api/devices", nil, c).Body.String()
	if !strings.Contains(raw, `"services":["_googlecast._tcp"]`) || !strings.Contains(raw, `"kind":"tv"`) {
		t.Fatal(raw)
	}
}
