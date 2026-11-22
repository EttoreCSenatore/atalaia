// Package web serves the compiled web app that is embedded in the binary.
//
// The app is built from web/ into internal/web/dist (npm run build). Until it
// is built, a short placeholder page explains how.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

const placeholder = `<!doctype html><meta charset="utf-8"><title>Atalaia</title>
<body style="font-family:system-ui;max-width:40rem;margin:4rem auto;padding:0 1rem">
<h1>Atalaia</h1><p>The web app is not built in this binary. Run <code>npm ci &amp;&amp; npm run build</code> in <code>web/</code>, then rebuild the program.</p>
<p>The API is available under <code>/api/</code>.</p>`

// Handler serves the embedded app; unknown paths fall back to index.html so the
// app can handle its own routes.
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		sub = dist
	}
	return handler(sub)
}

func handler(files fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if st, err := fs.Stat(files, name); err != nil || st.IsDir() {
			if name != "index.html" && path.Ext(name) != "" { // a missing asset is a 404, not the app
				http.NotFound(w, r)
				return
			}
			name = "index.html"
		}
		f, err := files.Open(name)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(placeholder))
			return
		}
		defer f.Close()
		st, _ := f.Stat()
		if path.Ext(name) == ".webmanifest" { // Go's mime table does not know it
			w.Header().Set("Content-Type", "application/manifest+json")
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // hashed file names
		} else {
			w.Header().Set("Cache-Control", "no-cache") // also sw.js, so a new build's service worker is seen at once
		}
		rs, ok := f.(readSeeker)
		if !ok {
			http.Error(w, "unreadable", http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, name, st.ModTime(), rs)
	})
}

type readSeeker interface {
	Read([]byte) (int, error)
	Seek(int64, int) (int64, error)
}
