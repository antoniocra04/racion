package web

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/antoniocra04/racion/internal/ration"
	"github.com/antoniocra04/racion/internal/store"
)

//go:embed templates/*.html static/*
var assets embed.FS

// designContract is trusted markup shipped with the application.
const designContract = template.HTML(`<!--
THESIS: Read the day's calories and macros at a glance.
OWN-WORLD: Off-white canvas, cobalt figures, one tinted summary and a plain list.
STORY: Check the active version, scan the day totals, then open the meal you need.
FIRST VIEWPORT: The selected plan is above a summary beside the list on desktop, above it on mobile.
FORM: Approved option 1, summary and list.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md
-->`)

var templates = template.Must(template.New("app").Funcs(template.FuncMap{
	"kcal":           calories,
	"grams":          grams,
	"id":             func(value int64) string { return strconv.FormatInt(value, 10) },
	"formNum":        formNumber,
	"designContract": func() template.HTML { return designContract },
}).ParseFS(assets, "templates/*.html"))

type app struct {
	store *store.Store
}

type page struct {
	Title    string
	Screen   string
	Error    string
	Plans    []ration.Plan
	Plan     *ration.Plan
	Products []ration.Product
	Query    string
	Product  productForm
	Draft    planDraft
	Inline   *inlineProduct
}

func New(db *store.Store) http.Handler {
	return newHandler(db, frontendFiles())
}

func newHandler(db *store.Store, frontend fs.FS) http.Handler {
	app := app{store: db}
	mux := http.NewServeMux()
	static, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	registerFrontend(mux, frontend, app)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /api/state", app.apiState)
	mux.HandleFunc("POST /api/plans", app.apiSavePlan)
	mux.HandleFunc("PUT /api/plans/{id}", app.apiSavePlan)
	mux.HandleFunc("DELETE /api/plans/{id}", app.apiDeletePlan)
	mux.HandleFunc("POST /api/plans/{id}/activate", app.apiActivatePlan)
	mux.HandleFunc("POST /api/products", app.apiSaveProduct)
	mux.HandleFunc("PUT /api/products/{id}", app.apiSaveProduct)
	mux.HandleFunc("DELETE /api/products/{id}", app.apiDeleteProduct)
	mux.HandleFunc("POST /plans/editor", app.updateEditor)
	mux.HandleFunc("POST /plans/activate", app.activate)
	mux.HandleFunc("POST /plans/{id}/delete", app.deletePlan)
	mux.HandleFunc("POST /products/save", app.saveProduct)
	mux.HandleFunc("POST /products/{id}/delete", app.deleteProduct)

	protected := http.NewCrossOriginProtection().Handler(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'; object-src 'none'")
		w.Header().Set("Referrer-Policy", "same-origin")
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		protected.ServeHTTP(w, r)
	})
}

func (a app) render(w http.ResponseWriter, r *http.Request, status int, data page) {
	if data.Screen == "plans" && data.Plans == nil {
		plans, err := a.store.Plans(r.Context())
		if err != nil {
			serverError(w, err)
			return
		}
		data.Plans = plans
	}
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "page", data); err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = out.WriteTo(w)
}

func (a app) fail(w http.ResponseWriter, r *http.Request, err error) {
	var message string
	status := http.StatusConflict
	switch {
	case errors.Is(err, store.ErrNotFound):
		status, message = http.StatusNotFound, "Страница не найдена."
	case errors.Is(err, store.ErrProductInUse):
		message = "Продукт используется в одном из рационов. Сначала убери его из этих рационов."
	case errors.Is(err, store.ErrActivePlan):
		message = "Сначала сделай активной другую версию рациона."
	default:
		serverError(w, err)
		return
	}
	a.render(w, r, status, page{Title: "Рацион", Screen: "error", Error: message})
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("request failed: %v", err)
	http.Error(w, "Не удалось выполнить запрос. Попробуй ещё раз.", http.StatusInternalServerError)
}

func parseForm(w http.ResponseWriter, r *http.Request, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Не удалось прочитать форму.", http.StatusBadRequest)
		return false
	}
	return true
}

func pathID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id
}

func planURL(id int64) string {
	return fmt.Sprintf("/plans/%d", id)
}

func calories(value float64) string {
	number := strconv.FormatFloat(math.Round(value), 'f', 0, 64)
	for i := len(number) - 3; i > 0; i -= 3 {
		number = number[:i] + "\u202f" + number[i:]
	}
	return number
}

func grams(value float64) string {
	return formNumber(math.Round(value*10) / 10)
}

func formNumber(value float64) string {
	return strings.ReplaceAll(strconv.FormatFloat(value, 'f', -1, 64), ".", ",")
}

func decimal(raw string) (float64, bool) {
	value := strings.ReplaceAll(strings.TrimSpace(raw), ",", ".")
	for _, ch := range value {
		if (ch < '0' || ch > '9') && ch != '.' {
			return 0, false
		}
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, len(value) <= 16 && err == nil && parsed <= 1_000_000
}
