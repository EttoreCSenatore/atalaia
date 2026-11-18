package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w
}

func TestServesAppAndFallsBackToIndex(t *testing.T) {
	h := handler(fstest.MapFS{
		"index.html":       {Data: []byte("<html>app</html>")},
		"assets/app-1.js":  {Data: []byte("console.log(1)")},
		"assets/app-1.css": {Data: []byte("body{}")},
	})
	if w := get(h, "/"); w.Code != 200 || !strings.Contains(w.Body.String(), "app") || w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("/: %d %q", w.Code, w.Header().Get("Cache-Control"))
	}
	if w := get(h, "/devices/abc"); w.Code != 200 || !strings.Contains(w.Body.String(), "app") {
		t.Fatalf("client route must get the app: %d", w.Code)
	}
	if w := get(h, "/assets/app-1.js"); w.Code != 200 || !strings.Contains(w.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("asset: %d %q", w.Code, w.Header().Get("Cache-Control"))
	}
	if w := get(h, "/assets/missing.js"); w.Code != 404 {
		t.Fatalf("missing asset must be 404, got %d", w.Code)
	}
	if w := get(h, "/../../etc/passwd"); w.Code != 200 || strings.Contains(w.Body.String(), "root:") {
		t.Fatalf("path traversal: %d", w.Code)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/", nil))
	if w.Code != 405 {
		t.Fatalf("POST: %d", w.Code)
	}
}

func TestPlaceholderWhenNotBuilt(t *testing.T) {
	h := handler(fstest.MapFS{"placeholder.txt": {Data: []byte("x")}})
	if w := get(h, "/"); w.Code != 200 || !strings.Contains(w.Body.String(), "not built") {
		t.Fatalf("placeholder: %d %q", w.Code, w.Body.String())
	}
}
