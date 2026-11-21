package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/antoniocra04/racion/internal/ration"
	"github.com/antoniocra04/racion/internal/store"
)

type nutritionJSON struct {
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Carbs   float64 `json:"carbs"`
	Kcal    float64 `json:"kcal"`
}

func nutritionView(n ration.Nutrition) nutritionJSON {
	return nutritionJSON{Protein: n.Protein, Fat: n.Fat, Carbs: n.Carbs, Kcal: n.Kcal()}
}

type productJSON struct {
	ID       int64         `json:"id"`
	Revision int64         `json:"revision"`
	Name     string        `json:"name"`
	State    string        `json:"state"`
	Per100   nutritionJSON `json:"per100"`
}

type itemJSON struct {
	ProductID    int64         `json:"productId"`
	ProductName  string        `json:"productName"`
	ProductState string        `json:"productState"`
	Grams        float64       `json:"grams"`
	Nutrition    nutritionJSON `json:"nutrition"`
}

type mealJSON struct {
	Name      string        `json:"name"`
	Items     []itemJSON    `json:"items"`
	Nutrition nutritionJSON `json:"nutrition"`
}

type planJSON struct {
	ID        int64         `json:"id"`
	Revision  int64         `json:"revision"`
	Name      string        `json:"name"`
	Active    bool          `json:"active"`
	Meals     []mealJSON    `json:"meals"`
	Nutrition nutritionJSON `json:"nutrition"`
}

func planView(plan ration.Plan) planJSON {
	out := planJSON{
		ID: plan.ID, Revision: plan.Revision, Name: plan.Name, Active: plan.Active,
		Meals: make([]mealJSON, 0, len(plan.Meals)), Nutrition: nutritionView(plan.Nutrition()),
	}
	for _, meal := range plan.Meals {
		view := mealJSON{Name: meal.Name, Items: make([]itemJSON, 0, len(meal.Items)), Nutrition: nutritionView(meal.Nutrition())}
		for _, item := range meal.Items {
			view.Items = append(view.Items, itemJSON{
				ProductID: item.Product.ID, ProductName: item.Product.Name, ProductState: item.Product.State,
				Grams: item.Grams, Nutrition: nutritionView(item.Nutrition()),
			})
		}
		out.Meals = append(out.Meals, view)
	}
	return out
}

func (a app) apiState(w http.ResponseWriter, r *http.Request) {
	plans, err := a.store.Plans(r.Context())
	if err != nil {
		apiStoreError(w, err)
		return
	}
	products, err := a.store.Products(r.Context())
	if err != nil {
		apiStoreError(w, err)
		return
	}
	sortProducts(products)
	response := struct {
		Plans    []planJSON    `json:"plans"`
		Products []productJSON `json:"products"`
	}{
		Plans: make([]planJSON, 0, len(plans)), Products: make([]productJSON, 0, len(products)),
	}
	for _, plan := range plans {
		response.Plans = append(response.Plans, planView(plan))
	}
	for _, p := range products {
		response.Products = append(response.Products, productJSON{
			ID: p.ID, Revision: p.Revision, Name: p.Name, State: p.State, Per100: nutritionView(p.Per100),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

type planInput struct {
	Revision int64  `json:"revision"`
	Name     string `json:"name"`
	Meals    []struct {
		Name  string `json:"name"`
		Items []struct {
			ProductID int64  `json:"productId"`
			Grams     string `json:"grams"`
		} `json:"items"`
	} `json:"meals"`
}

func (a app) apiSavePlan(w http.ResponseWriter, r *http.Request) {
	var input planInput
	if !readJSON(w, r, &input) {
		return
	}
	id, ok := apiSaveID(w, r, input.Revision)
	if !ok {
		return
	}
	if len(input.Meals) > maxMeals {
		apiError(w, http.StatusUnprocessableEntity, "Слишком много приёмов пищи.", nil)
		return
	}
	draft := planDraft{ID: id, Revision: input.Revision, Name: input.Name}
	for _, meal := range input.Meals {
		if len(meal.Items) > maxItems {
			apiError(w, http.StatusUnprocessableEntity, "Слишком много продуктов в одном приёме пищи.", nil)
			return
		}
		form := mealDraft{Name: meal.Name}
		for _, item := range meal.Items {
			form.Items = append(form.Items, itemDraft{ProductID: strconv.FormatInt(item.ProductID, 10), Grams: item.Grams})
		}
		draft.Meals = append(draft.Meals, form)
	}
	plan, valid := draft.validate()
	if !valid {
		fields := make(map[string]string)
		if draft.NameError != "" {
			fields["name"] = draft.NameError
		}
		for mi, meal := range draft.Meals {
			if meal.Error != "" {
				fields[fmt.Sprintf("meal.%d.name", mi)] = meal.Error
			}
			for ii, item := range meal.Items {
				if item.ProductError != "" {
					fields[fmt.Sprintf("meal.%d.%d.productId", mi, ii)] = item.ProductError
				}
				if item.WeightError != "" {
					fields[fmt.Sprintf("meal.%d.%d.grams", mi, ii)] = item.WeightError
				}
			}
		}
		apiError(w, http.StatusUnprocessableEntity, "Проверь отмеченные поля.", fields)
		return
	}
	saved, err := a.store.SavePlan(r.Context(), plan)
	if err != nil {
		apiStoreError(w, err)
		return
	}
	writeSaved(w, id, saved)
}

type productInput struct {
	Revision int64  `json:"revision"`
	Token    string `json:"token"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Protein  string `json:"protein"`
	Fat      string `json:"fat"`
	Carbs    string `json:"carbs"`
}

func (a app) apiSaveProduct(w http.ResponseWriter, r *http.Request) {
	var input productInput
	if !readJSON(w, r, &input) {
		return
	}
	id, ok := apiSaveID(w, r, input.Revision)
	if !ok {
		return
	}
	p, form, valid := productFromValues(url.Values{
		"name": {input.Name}, "state": {input.State}, "protein": {input.Protein},
		"fat": {input.Fat}, "carbs": {input.Carbs}, "product_token": {input.Token},
	}, "name")
	if !valid {
		for _, field := range form.Numbers {
			if field.Error != "" {
				form.Errors[field.Key] = field.Error
			}
		}
		apiError(w, http.StatusUnprocessableEntity, "Проверь отмеченные поля.", form.Errors)
		return
	}
	p.ID, p.Revision = id, input.Revision
	saved, err := a.store.SaveProduct(r.Context(), p, form.Token)
	if err != nil {
		apiStoreError(w, err)
		return
	}
	writeSaved(w, id, saved)
}

func (a app) apiDeletePlan(w http.ResponseWriter, r *http.Request) {
	finishMutation(w, a.store.DeletePlan(r.Context(), pathID(r)))
}

func (a app) apiActivatePlan(w http.ResponseWriter, r *http.Request) {
	finishMutation(w, a.store.ActivatePlan(r.Context(), pathID(r)))
}

func (a app) apiDeleteProduct(w http.ResponseWriter, r *http.Request) {
	finishMutation(w, a.store.DeleteProduct(r.Context(), pathID(r)))
}

func finishMutation(w http.ResponseWriter, err error) {
	if err != nil {
		apiStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiSaveID(w http.ResponseWriter, r *http.Request, revision int64) (int64, bool) {
	if r.Method == http.MethodPost {
		return 0, true
	}
	if id := pathID(r); id > 0 && revision > 0 {
		return id, true
	}
	apiError(w, http.StatusBadRequest, "Некорректный номер или версия записи.", nil)
	return 0, false
}

func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		apiError(w, http.StatusUnsupportedMediaType, "Ожидался запрос в формате JSON.", nil)
		return false
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil || d.Decode(new(any)) != io.EOF {
		apiError(w, http.StatusBadRequest, "Не удалось прочитать запрос.", nil)
		return false
	}
	return true
}

func apiStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		apiError(w, http.StatusNotFound, "Страница не найдена.", nil)
	case errors.Is(err, store.ErrProductInUse):
		apiError(w, http.StatusConflict, "Продукт используется в рационе. Сначала убери его из рациона.", nil)
	case errors.Is(err, store.ErrProductMissing):
		apiError(w, http.StatusUnprocessableEntity, "Один из продуктов удалён. Выбери его замену.", nil)
	case errors.Is(err, store.ErrProductConflict):
		apiError(w, http.StatusConflict, "Карточку изменили или удалили. Открой актуальную карточку в другой вкладке и перенеси свои правки.", nil)
	case errors.Is(err, store.ErrActivePlan):
		apiError(w, http.StatusConflict, "Сначала сделай активной другую версию рациона.", nil)
	case errors.Is(err, store.ErrConflict):
		apiError(w, http.StatusConflict, "Рацион изменился. Открой актуальную версию в другой вкладке и перенеси свои правки.", nil)
	default:
		log.Printf("api request failed: %v", err)
		apiError(w, http.StatusInternalServerError, "Не удалось выполнить запрос. Попробуй ещё раз.", nil)
	}
}

func apiError(w http.ResponseWriter, status int, message string, fields map[string]string) {
	writeJSON(w, status, struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields,omitempty"`
	}{Error: message, Fields: fields})
}

func writeSaved(w http.ResponseWriter, oldID, id int64) {
	status := http.StatusOK
	if oldID == 0 {
		status = http.StatusCreated
	}
	writeJSON(w, status, struct {
		ID int64 `json:"id"`
	}{ID: id})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON: %v", err)
	}
}
