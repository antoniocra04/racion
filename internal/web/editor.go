package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/antoniocra04/racion/internal/ration"
	"github.com/antoniocra04/racion/internal/store"
)

const (
	maxMeals = 50
	maxItems = 100
)

type planDraft struct {
	ID, Revision int64
	Name         string
	NameError    string
	Meals        []mealDraft
}

type mealDraft struct {
	Name  string
	Error string
	Items []itemDraft
}

type itemDraft struct {
	ProductID string
	Grams     string
	Error     string
}

type inlineProduct struct {
	Meal, Item int
	Form       productForm
}

func (a app) openEditor(w http.ResponseWriter, r *http.Request) {
	plans, err := a.store.Plans(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	data := page{Screen: "editor", Plans: plans}
	if r.PathValue("id") == "" {
		data.Draft.Meals = []mealDraft{{Name: "Завтрак"}}
	} else {
		plan := findPlan(plans, pathID(r))
		if plan == nil {
			a.fail(w, r, store.ErrNotFound)
			return
		}
		data.Draft = draftFromPlan(*plan)
		if strings.HasSuffix(r.URL.Path, "/copy") {
			data.Draft.ID, data.Draft.Revision = 0, 0
			data.Draft.Name = string([]rune(plan.Name)[:min(92, utf8.RuneCountInString(plan.Name))]) + " — копия"
		}
	}
	a.renderEditor(w, r, http.StatusOK, data)
}

func draftFromPlan(plan ration.Plan) planDraft {
	draft := planDraft{ID: plan.ID, Revision: plan.Revision, Name: plan.Name}
	for _, meal := range plan.Meals {
		current := mealDraft{Name: meal.Name}
		for _, item := range meal.Items {
			current.Items = append(current.Items, itemDraft{
				ProductID: strconv.FormatInt(item.Product.ID, 10),
				Grams:     formNumber(item.Grams),
			})
		}
		draft.Meals = append(draft.Meals, current)
	}
	return draft
}

func (a app) renderEditor(w http.ResponseWriter, r *http.Request, status int, data page) {
	products, err := a.store.Products(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	sortProducts(products)
	data.Products = products
	data.Screen, data.Title = "editor", "Создать рацион"
	if data.Draft.ID != 0 {
		data.Title = "Изменить рацион"
	}
	a.render(w, r, status, data)
}

func (a app) updateEditor(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r, 256<<10) {
		return
	}
	draft, ok := readDraft(r.PostForm)
	if !ok {
		http.Error(w, "Не удалось прочитать состав рациона.", http.StatusBadRequest)
		return
	}
	data := page{Draft: draft}
	action, position, _ := strings.Cut(r.PostForm.Get("intent"), ":")
	mealText, itemText, _ := strings.Cut(position, ":")
	mealIndex, mealOK := indexOf(mealText, len(draft.Meals))
	status := http.StatusOK

	switch {
	case action == "save" || action == "":
		a.saveDraft(w, r, &data)
		return
	case action == "add-meal" && len(draft.Meals) < maxMeals:
		data.Draft.Meals = append(data.Draft.Meals, mealDraft{})
	case action == "remove-meal" && mealOK:
		data.Draft.Meals = remove(draft.Meals, mealIndex)
	case action == "cancel-product":
		// The product was never written. The rest of the draft stays intact.
	case action == "add-item" && mealOK && len(draft.Meals[mealIndex].Items) < maxItems:
		meal := &data.Draft.Meals[mealIndex]
		meal.Items = append(meal.Items, itemDraft{})
	case mealOK:
		meal := &data.Draft.Meals[mealIndex]
		itemIndex, ok := indexOf(itemText, len(meal.Items))
		if !ok {
			http.Error(w, "Позиция не найдена.", http.StatusBadRequest)
			return
		}
		switch action {
		case "remove-item":
			meal.Items = remove(meal.Items, itemIndex)
		case "new-product":
			data.Inline = &inlineProduct{Meal: mealIndex, Item: itemIndex, Form: newProductForm("product_name")}
		case "save-product":
			product, form, valid := readProduct(r, "product_name")
			if !valid {
				data.Inline = &inlineProduct{Meal: mealIndex, Item: itemIndex, Form: form}
				status = http.StatusUnprocessableEntity
				break
			}
			id, err := a.store.SaveProduct(r.Context(), product, form.Token)
			if err != nil {
				a.fail(w, r, err)
				return
			}
			meal.Items[itemIndex].ProductID = strconv.FormatInt(id, 10)
		default:
			http.Error(w, "Неизвестное действие.", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Не удалось изменить состав рациона.", http.StatusBadRequest)
		return
	}
	a.renderEditor(w, r, status, data)
}

func readDraft(form url.Values) (planDraft, bool) {
	draft := planDraft{Name: form.Get("name")}
	if id := form.Get("id"); id != "" && id != "0" {
		var err error
		draft.ID, err = strconv.ParseInt(id, 10, 64)
		if err != nil || draft.ID < 1 {
			return draft, false
		}
		draft.Revision, err = strconv.ParseInt(form.Get("revision"), 10, 64)
		if err != nil || draft.Revision < 1 {
			return draft, false
		}
	}
	names := form["meal_name"]
	if len(names) > maxMeals {
		return draft, false
	}
	for i, name := range names {
		meal := mealDraft{Name: name}
		products := form[fmt.Sprintf("product_%d", i)]
		weights := form[fmt.Sprintf("grams_%d", i)]
		if len(products) != len(weights) || len(products) > maxItems {
			return draft, false
		}
		for j, product := range products {
			meal.Items = append(meal.Items, itemDraft{ProductID: product, Grams: weights[j]})
		}
		draft.Meals = append(draft.Meals, meal)
	}
	return draft, true
}

func (a app) saveDraft(w http.ResponseWriter, r *http.Request, data *page) {
	plan, valid := data.Draft.validate()
	if !valid {
		data.Error = "Проверь отмеченные поля."
		a.renderEditor(w, r, http.StatusUnprocessableEntity, *data)
		return
	}
	id, err := a.store.SavePlan(r.Context(), plan)
	switch {
	case errors.Is(err, store.ErrProductMissing):
		data.Error = "Один из продуктов удалён. Выбери его замену и сохрани рацион снова."
		a.renderEditor(w, r, http.StatusUnprocessableEntity, *data)
	case errors.Is(err, store.ErrConflict):
		data.Error = "Рацион изменился с момента открытия. Открой актуальную версию в другой вкладке и перенеси свои правки."
		a.renderEditor(w, r, http.StatusConflict, *data)
	case err != nil:
		a.fail(w, r, err)
	default:
		http.Redirect(w, r, planURL(id), http.StatusSeeOther)
	}
}

func (d *planDraft) validate() (ration.Plan, bool) {
	plan := ration.Plan{ID: d.ID, Revision: d.Revision, Name: strings.TrimSpace(d.Name)}
	valid := nameOK(plan.Name)
	if !valid {
		d.NameError = "Укажи название рациона длиной от 1 до 100 символов."
	}
	for i := range d.Meals {
		draftMeal := &d.Meals[i]
		meal := ration.Meal{Name: strings.TrimSpace(draftMeal.Name)}
		if !nameOK(meal.Name) {
			draftMeal.Error = "Укажи название приёма пищи длиной от 1 до 100 символов."
			valid = false
		}
		for j := range draftMeal.Items {
			item := &draftMeal.Items[j]
			id, err := strconv.ParseInt(item.ProductID, 10, 64)
			amount, ok := decimal(item.Grams)
			switch {
			case err != nil || id < 1:
				item.Error = "Выбери продукт."
			case !ok || amount <= 0:
				item.Error = "Укажи вес больше 0 и не больше 1 000 000 г."
			}
			valid = valid && item.Error == ""
			meal.Items = append(meal.Items, ration.Item{Product: ration.Product{ID: id}, Grams: amount})
		}
		plan.Meals = append(plan.Meals, meal)
	}
	return plan, valid
}

func nameOK(value string) bool {
	return value != "" && utf8.RuneCountInString(value) <= 100
}

func indexOf(value string, size int) (int, bool) {
	index, err := strconv.Atoi(value)
	return index, err == nil && index >= 0 && index < size
}

func remove[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}
