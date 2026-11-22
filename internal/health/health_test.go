package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsLow(t *testing.T) {
	const g = 1 << 30
	for _, c := range []struct {
		free, total uint64
		want        bool
	}{
		{100 * g, 200 * g, false},
		{4 * g, 100 * g, true},    // under 5%
		{400 << 20, 2 * g, true},  // under 512 MiB
		{600 << 20, 1 * g, false}, // 58% free and over 512 MiB
		{0, 0, false},             // unknown
	} {
		if got := IsLow(c.free, c.total); got != c.want {
			t.Errorf("IsLow(%d, %d) = %v", c.free, c.total, got)
		}
	}
}

func TestDiskFor(t *testing.T) {
	d, err := DiskFor(t.TempDir())
	if err != nil {
		t.Skip(err)
	}
	if d.Total == 0 || d.Free > d.Total {
		t.Fatalf("odd disk: %+v", d)
	}
}

func TestHeartbeat(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/elsewhere", http.StatusFound)
		}
	}))
	defer srv.Close()
	h, err := ParseHeartbeat(srv.URL + "/ping/abc/")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Ping(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	if err := h.Ping(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	if strings.Join(paths, " ") != "/ping/abc /ping/abc/fail" {
		t.Fatalf("paths = %v", paths)
	}
	r, _ := ParseHeartbeat(srv.URL + "/redirect")
	if err := r.Ping(context.Background(), true); err == nil {
		t.Fatal("redirect was followed or accepted")
	}
	down, _ := ParseHeartbeat("http://127.0.0.1:1/secret-uuid")
	if err := down.Ping(context.Background(), true); err == nil || strings.Contains(err.Error(), "secret-uuid") {
		t.Fatalf("error leaks the URL or is missing: %v", err)
	}
	var off *Heartbeat
	if err := off.Ping(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"ftp://x/y", "not a url", "https://"} {
		if _, err := ParseHeartbeat(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	if h, err := ParseHeartbeat("  "); h != nil || err != nil {
		t.Fatal("empty should be off")
	}
}

func TestMetrics(t *testing.T) {
	var m Metrics
	m.Add("atalaia_build_info", "gauge", "Build.", 1, "version", "v1\"x\\\n\x01")
	m.Add("atalaia_devices", "gauge", "Devices.", 12)
	want := "# HELP atalaia_build_info Build.\n# TYPE atalaia_build_info gauge\natalaia_build_info{version=\"v1\\\"x\\\\\\n\"} 1\n" +
		"# HELP atalaia_devices Devices.\n# TYPE atalaia_devices gauge\natalaia_devices 12\n"
	if m.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", m.String(), want)
	}
}
