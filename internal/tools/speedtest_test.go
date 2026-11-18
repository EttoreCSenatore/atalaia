package tools

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSpeedTestMeasuresAndCapsTheDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("x"), 2<<20))
	}))
	defer ts.Close()
	r, err := SpeedTest(context.Background(), nil, ts.URL, 1<<20, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if r.Bytes != 1<<20 || r.Mbps <= 0 || r.Server != "127.0.0.1" || r.LatencyMs < 0 {
		t.Fatalf("%+v", r)
	}
}

func TestSpeedTestRefusesRedirectsAndBadAnswers(t *testing.T) {
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/", 302)
	}))
	defer redir.Close()
	if _, err := SpeedTest(context.Background(), nil, redir.URL, 1<<20, time.Second); err == nil || strings.Contains(err.Error(), "169.254") {
		t.Fatalf("a redirect must not be followed: %v", err)
	}
	tiny := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("hi")) }))
	defer tiny.Close()
	if _, err := SpeedTest(context.Background(), nil, tiny.URL, 1<<20, time.Second); err == nil {
		t.Fatal("too little data must not give a number")
	}
	for _, bad := range []string{"", "ftp://x/y", "file:///etc/passwd", "://"} {
		if _, err := SpeedTest(context.Background(), nil, bad, 1<<20, time.Second); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestSpeedTestStopsAtTheTimeLimit(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fl := w.(http.Flusher)
		for i := 0; i < 50; i++ {
			_, _ = w.Write(bytes.Repeat([]byte("x"), 4096))
			fl.Flush()
			select {
			case <-r.Context().Done():
				return
			case <-time.After(50 * time.Millisecond):
			}
		}
	}))
	defer slow.Close()
	start := time.Now()
	r, err := SpeedTest(context.Background(), nil, slow.URL, 100<<20, 400*time.Millisecond)
	if time.Since(start) > 2*time.Second {
		t.Fatal("did not stop at the time limit")
	}
	if err != nil || r.Bytes < 4096 {
		t.Fatalf("%+v %v", r, err)
	}
}
