package web

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antoniocra04/racion/internal/store"
)

func TestSPAAPIWorkflow(t *testing.T) {
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "ration.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	handler := New(db)
	do := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			r.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s returned %d, want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		return w
	}
	getID := func(w *httptest.ResponseRecorder) int64 {
		t.Helper()
		var saved struct{ ID int64 }
		if err := json.NewDecoder(w.Body).Decode(&saved); err != nil || saved.ID < 1 {
			t.Fatalf("save did not return a valid ID: %v, %v", saved, err)
		}
		return saved.ID
	}

	do(http.MethodGet, "/healthz", "", http.StatusOK)
	do(http.MethodPost, "/api/products", `{"name":"","protein":"not a number","fat":"0","carbs":"0"}`, http.StatusUnprocessableEntity)
	product := `{"name":"Овсянка","state":"сухой","protein":"12,5","fat":"7","carbs":"60","token":"test-create-oats"}`
	productID := getID(do(http.MethodPost, "/api/products", product, http.StatusCreated))
	if retryID := getID(do(http.MethodPost, "/api/products", product, http.StatusCreated)); retryID != productID {
		t.Fatalf("retry created a second product: %d, %d", productID, retryID)
	}
	plan := fmt.Sprintf(`{"name":"Обычный день","meals":[{"name":"Завтрак","items":[{"productId":%d,"grams":"80,0"}]}]}`, productID)
	planID := getID(do(http.MethodPost, "/api/plans", plan, http.StatusCreated))
	do(http.MethodPost, fmt.Sprintf("/api/plans/%d/activate", planID), "", http.StatusNoContent)

	verifyCalories := func(want float64) {
		t.Helper()
		var state struct {
			Plans []struct {
				ID        int64
				Active    bool
				Nutrition struct{ Kcal float64 }
			}
		}
		response := do(http.MethodGet, "/api/state", "", http.StatusOK)
		if err := json.NewDecoder(response.Body).Decode(&state); err != nil {
			t.Fatal(err)
		}
		if len(state.Plans) != 1 || state.Plans[0].ID != planID || !state.Plans[0].Active || math.Abs(state.Plans[0].Nutrition.Kcal-want) > 1e-8 {
			t.Fatalf("want active plan %d with %.1f kcal: %+v", planID, want, state)
		}
	}
	verifyCalories(282.4)
	path := fmt.Sprintf("/api/products/%d", productID)
	revised := `{"revision":1,"name":"Овсянка","state":"сухой","protein":"15","fat":"7","carbs":"60"}`
	do(http.MethodPut, path, revised, http.StatusOK)
	do(http.MethodPut, path, revised, http.StatusConflict)
	verifyCalories(290.4)
	do(http.MethodDelete, path, "", http.StatusConflict)
	do(http.MethodDelete, fmt.Sprintf("/api/plans/%d", planID), "", http.StatusNoContent)
	do(http.MethodDelete, path, "", http.StatusNoContent)
}
