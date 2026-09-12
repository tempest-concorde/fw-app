package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// distFS embeds the built React + PatternFly single-page app. The `dist`
// directory is produced by `npm run build` in this directory; a placeholder
// index.html is committed so the tree builds even before the frontend build.
//
//go:embed dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded SPA with a
// single-page-app fallback: any request path that does not map to a real file
// is served index.html so client-side routes (e.g. /app, /login) resolve.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			// Unknown non-file path: fall back to index.html for client routing.
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
