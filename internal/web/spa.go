package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:spa
var bundledFrontend embed.FS

func frontendFiles() fs.FS {
	root, err := fs.Sub(bundledFrontend, "spa/generated")
	if err != nil {
		return nil
	}
	if _, err := fs.Stat(root, "index.html"); err != nil {
		return nil
	}
	return root
}

func registerFrontend(mux *http.ServeMux, frontend fs.FS, app app) {
	pages := []struct {
		route  string
		legacy http.HandlerFunc
	}{
		{"GET /{$}", app.viewRation},
		{"GET /plans", app.listPlans},
		{"GET /plans/{id}", app.viewRation},
		{"GET /plans/new", app.openEditor},
		{"GET /plans/{id}/edit", app.openEditor},
		{"GET /plans/{id}/copy", app.openEditor},
		{"GET /products", app.listProducts},
		{"GET /products/new", app.openProduct},
		{"GET /products/{id}/edit", app.openProduct},
	}
	for _, page := range pages {
		handler := page.legacy
		if frontend != nil {
			handler = func(w http.ResponseWriter, r *http.Request) {
				http.ServeFileFS(w, r, frontend, "index.html")
			}
		}
		mux.HandleFunc(page.route, handler)
	}
	if frontend != nil {
		mux.HandleFunc("GET /assets/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := "assets/" + r.PathValue("name")
			if !fs.ValidPath(name) || strings.Contains(r.PathValue("name"), "/") {
				http.NotFound(w, r)
				return
			}
			if info, err := fs.Stat(frontend, name); err != nil || !info.Mode().IsRegular() {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			http.ServeFileFS(w, r, frontend, name)
		})
	}
}
