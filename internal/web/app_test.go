package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antoniocra04/racion/internal/store"
	"github.com/antoniocra04/racion/internal/web"
)

func TestRationWorkflow(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "ration.db")
	db, err := store.Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	handler := web.New(db)
	check(t, request(handler, "GET", "/", nil), 200, "Создать рацион")

	product := url.Values{
		"name": {"Курица"}, "state": {"готовый"}, "kcal": {"120"}, "protein": {"10"},
		"fat": {"4"}, "carbs": {"11,0"},
	}
	check(t, request(handler, "POST", "/products/save", product), 303, "")
	check(t, request(handler, "GET", "/products?q=КУРИЦА", nil), 200, "Курица")

	draft := url.Values{
		"intent": {"save"}, "name": {"Основной"}, "meal_name": {"Обед"},
		"product_0": {"1"}, "grams_0": {"150"},
	}
	check(t, request(handler, "POST", "/plans/editor", draft), 303, "")
	daily := request(handler, "GET", "/", nil)
	check(t, daily, 200, "180", "16,5", "Основной", "150 г")
	_, meal, found := strings.Cut(daily.Body.String(), `<details class="meal">`)
	if !found {
		t.Fatal("the meal should start collapsed")
	}
	summary, _, found := strings.Cut(meal, "</summary>")
	if !found || !strings.Contains(summary, "Обед") || !strings.Contains(summary, "180") ||
		strings.Contains(summary, "Белки") || strings.Contains(summary, "16,5") {
		t.Fatalf("the collapsed meal should show its name and calories only: %s", summary)
	}
	check(t, request(handler, "GET", "/plans/1/copy", nil), 200, "150", "Основной — копия")

	draft.Set("name", "Тренировочный")
	draft.Set("grams_0", "200")
	check(t, request(handler, "POST", "/plans/editor", draft), 303, "")
	check(t, request(handler, "GET", "/", nil), 200, "180", "Основной")

	product.Set("id", "1")
	product.Set("revision", "1")
	product.Set("kcal", "140")
	check(t, request(handler, "POST", "/products/save", product), 303, "")
	product.Set("kcal", "120")
	check(t, request(handler, "POST", "/products/save", product), 409, "120")
	check(t, request(handler, "GET", "/plans/1", nil), 200, "210", "150 г")
	check(t, request(handler, "GET", "/plans/2", nil), 200, "280", "200 г")
	check(t, request(handler, "POST", "/plans/activate", url.Values{"id": {"2"}}), 303, "")
	check(t, request(handler, "POST", "/plans/2/delete", nil), 409, "")
	check(t, request(handler, "POST", "/products/1/delete", nil), 409, "")

	backup := filepath.Join(t.TempDir(), "backup.db")
	if err := db.Backup(ctx, backup); err != nil {
		t.Fatal(err)
	}
	copy, err := store.Open(ctx, backup)
	if err != nil {
		t.Fatal(err)
	}
	defer copy.Close()
	check(t, request(web.New(copy), "GET", "/", nil), 200, "Тренировочный", "280", "Курица")
}

func TestEditorKeepsDraftAndAddsProduct(t *testing.T) {
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "ration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := web.New(db)
	draft := url.Values{
		"intent": {"new-product:0:0"}, "name": {"Основной"}, "meal_name": {"Завтрак"},
		"product_0": {""}, "grams_0": {"125,5"},
	}
	check(t, request(handler, "POST", "/plans/editor", draft), 200, "125,5", "Завтрак", "Новый продукт")
	draft.Set("intent", "save-product:0:0")
	draft.Set("product_token", "P4I42N5RO7VRPAJ6XK64VP7MRY")
	draft.Set("product_name", "Рис")
	draft.Set("kcal", "100")
	draft.Set("protein", "0")
	draft.Set("fat", "-1")
	draft.Set("carbs", "15")
	check(t, request(handler, "POST", "/plans/editor", draft), 422, "125,5", "Завтрак", "Рис")
	draft.Set("fat", "0")
	check(t, request(handler, "POST", "/plans/editor", draft), 200, "125,5", "Рис", "selected")
	check(t, request(handler, "POST", "/plans/editor", draft), 200, "125,5", "Рис", "selected")
	catalog := request(handler, "GET", "/products", nil)
	if strings.Count(catalog.Body.String(), `class="record-title"`) != 1 {
		t.Fatal("resubmitting the inline form created a second product")
	}
	draft.Set("intent", "save")
	draft.Set("product_0", "999")
	check(t, request(handler, "POST", "/plans/editor", draft), 422, "125,5")
	check(t, request(handler, "GET", "/", nil), 200, "Создать рацион")
}

func request(handler http.Handler, method, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func check(t *testing.T, recorder *httptest.ResponseRecorder, status int, fragments ...string) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, status, recorder.Body)
	}
	for _, fragment := range fragments {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Errorf("missing %q in %s", fragment, recorder.Body)
		}
	}
}
