package web_test

import (
	"context"
	"database/sql"
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

	form := request(handler, "GET", "/products/new", nil)
	check(t, form, 200, "Расчётная калорийность")
	if strings.Contains(form.Body.String(), `name="kcal"`) {
		t.Fatal("calories should not be an editable field")
	}
	product := url.Values{
		"name": {"Курица"}, "state": {"готовый"}, "protein": {"10"},
		"fat": {"4"}, "carbs": {"11,0"},
	}
	check(t, request(handler, "POST", "/products/save", product), 303, "")

	// Simulate a product saved by the old version with a calorie value from the label.
	legacy, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer legacy.Close()
	if _, err := legacy.ExecContext(ctx, "UPDATE products SET kcal = 900 WHERE id = 1"); err != nil {
		t.Fatal(err)
	}
	check(t, request(handler, "GET", "/products?q=КУРИЦА", nil), 200, "Курица", "120 ккал")

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
	versions := request(handler, "GET", "/plans", nil)
	check(t, versions, 200, "Сделать активным", "Тренировочный")
	if strings.Contains(versions.Body.String(), "data-switcher") {
		t.Fatal("the versions list should use its own activation buttons")
	}
	edit := request(handler, "GET", "/plans/1/edit", nil)
	if strings.Count(edit.Body.String(), ">Сохранить рацион</button>") != 1 ||
		strings.Contains(edit.Body.String(), "data-switcher") {
		t.Fatal("the editor needs one save action and no version switch")
	}

	product.Set("id", "1")
	product.Set("revision", "1")
	product.Set("protein", "15")
	product.Set("kcal", "999")
	check(t, request(handler, "POST", "/products/save", product), 303, "")
	product.Set("protein", "10")
	check(t, request(handler, "POST", "/products/save", product), 409, `name="protein" inputmode="decimal" value="10"`, "120")
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
	newPlan := request(handler, "GET", "/plans/new", nil)
	if strings.Count(newPlan.Body.String(), ">Сохранить рацион</button>") != 1 {
		t.Fatal("the create screen needs exactly one save action")
	}
	check(t, newPlan, 200, `value="add-new-product:0"`, "Создать первый продукт")

	firstProduct := url.Values{"intent": {"add-new-product:0"}, "name": {"Основной"}, "meal_name": {"Завтрак"}}
	inline := request(handler, "POST", "/plans/editor", firstProduct)
	check(t, inline, 200, `value="save-product:0:0"`, `name="inline_added_item" value="0:0"`, "Расчётная калорийность")
	if strings.Count(inline.Body.String(), `value="save-product:0:0"`) != 1 ||
		strings.Contains(inline.Body.String(), ">Сохранить рацион</button>") {
		t.Fatal("an open product form should have only its own save action")
	}
	firstProduct.Set("product_0", "")
	firstProduct.Set("grams_0", "")
	firstProduct.Set("inline_added_item", "0:0")
	firstProduct.Set("intent", "cancel-product:0:0")
	back := request(handler, "POST", "/plans/editor", firstProduct)
	check(t, back, 200, "Основной", "Завтрак", "Создать первый продукт")
	if strings.Contains(back.Body.String(), `name="product_0"`) {
		t.Fatal("cancelling creation of the first product should leave no blank row")
	}

	draft := url.Values{
		"intent": {"new-product:0:0"}, "name": {"Основной"}, "meal_name": {"Завтрак"},
		"product_0": {""}, "grams_0": {"125,5"},
	}
	check(t, request(handler, "POST", "/plans/editor", draft), 200, "125,5", "Завтрак", "Новый продукт")
	draft.Set("intent", "save-product:0:0")
	draft.Set("product_token", "P4I42N5RO7VRPAJ6XK64VP7MRY")
	draft.Set("product_name", "Рис")
	draft.Set("protein", "0")
	draft.Set("fat", "-1")
	draft.Set("carbs", "15")
	check(t, request(handler, "POST", "/plans/editor", draft), 422, "125,5", "Завтрак", "Рис")
	draft.Set("fat", "0")
	check(t, request(handler, "POST", "/plans/editor", draft), 200, "125,5", "Рис", "selected")
	check(t, request(handler, "POST", "/plans/editor", draft), 200, "125,5", "Рис", "selected")
	catalog := request(handler, "GET", "/products", nil)
	check(t, catalog, 200, "60 ккал")
	if strings.Count(catalog.Body.String(), `class="record-title"`) != 1 {
		t.Fatal("resubmitting the inline form created a second product")
	}
	if strings.Count(catalog.Body.String(), `href="/products/1/edit"`) != 1 {
		t.Fatal("each product should have a single entry point to its editor")
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
