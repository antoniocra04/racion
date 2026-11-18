package web

import (
	"crypto/rand"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/antoniocra04/racion/internal/ration"
	"github.com/antoniocra04/racion/internal/store"
)

type productForm struct {
	ID        int64
	Revision  int64
	Token     string
	NameField string
	Name      string
	State     string
	Numbers   []numberField
	Energy    string
	Errors    map[string]string
}

type numberField struct {
	Key         string
	Label       string
	KcalPerGram int
	Value       string
	Error       string
}

func newProductForm(nameField string) productForm {
	return productForm{
		NameField: nameField,
		Token:     rand.Text(),
		Numbers: []numberField{
			{Key: "protein", Label: "Белки, г", KcalPerGram: ration.ProteinKcalPerGram},
			{Key: "fat", Label: "Жиры, г", KcalPerGram: ration.FatKcalPerGram},
			{Key: "carbs", Label: "Углеводы, г", KcalPerGram: ration.CarbsKcalPerGram},
		},
		Energy: "—",
		Errors: make(map[string]string),
	}
}

func (a app) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := a.store.Products(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	folded := strings.ToLower(query)
	var filtered []ration.Product
	for _, product := range products {
		if strings.Contains(strings.ToLower(product.Name), folded) {
			filtered = append(filtered, product)
		}
	}
	sortProducts(filtered)
	a.render(w, r, http.StatusOK, page{Title: "Мои продукты", Screen: "products", Query: query, Products: filtered})
}

func sortProducts(products []ration.Product) {
	sort.SliceStable(products, func(i, j int) bool {
		return strings.ToLower(products[i].Name) < strings.ToLower(products[j].Name)
	})
}

func (a app) openProduct(w http.ResponseWriter, r *http.Request) {
	form := newProductForm("name")
	title := "Добавить продукт"
	if r.PathValue("id") != "" {
		product, err := a.store.Product(r.Context(), pathID(r))
		if err != nil {
			a.fail(w, r, err)
			return
		}
		title = "Изменить продукт"
		form.ID, form.Revision, form.Name, form.State = product.ID, product.Revision, product.Name, product.State
		values := []float64{product.Per100.Protein, product.Per100.Fat, product.Per100.Carbs}
		for i, value := range values {
			form.Numbers[i].Value = formNumber(value)
		}
		form.Energy = calories(product.Per100.Kcal())
	}
	a.render(w, r, http.StatusOK, page{Title: title, Screen: "product", Product: form})
}

func (a app) saveProduct(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r, 16<<10) {
		return
	}
	product, form, valid := readProduct(r, "name")
	title := "Добавить продукт"
	if value := r.PostForm.Get("id"); value != "" && value != "0" {
		id, err := strconv.ParseInt(value, 10, 64)
		revision, revisionErr := strconv.ParseInt(r.PostForm.Get("revision"), 10, 64)
		if err != nil || id < 1 || revisionErr != nil || revision < 1 {
			http.Error(w, "Некорректный номер продукта.", http.StatusBadRequest)
			return
		}
		product.ID, form.ID = id, id
		product.Revision, form.Revision = revision, revision
		title = "Изменить продукт"
	}
	if !valid {
		a.render(w, r, http.StatusUnprocessableEntity, page{Title: title, Screen: "product", Product: form})
		return
	}
	if _, err := a.store.SaveProduct(r.Context(), product, form.Token); err != nil {
		if errors.Is(err, store.ErrProductConflict) {
			a.render(w, r, http.StatusConflict, page{
				Title: title, Screen: "product", Product: form,
				Error: "Карточку изменили или удалили. Открой актуальную карточку в другой вкладке и перенеси свои правки.",
			})
			return
		}
		a.fail(w, r, err)
		return
	}
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func readProduct(r *http.Request, nameField string) (ration.Product, productForm, bool) {
	form := newProductForm(nameField)
	if key := r.PostForm.Get("product_token"); key != "" && len(key) <= 128 {
		form.Token = key
	}
	form.Name = r.PostForm.Get(nameField)
	form.State = r.PostForm.Get("state")
	product := ration.Product{Name: strings.TrimSpace(form.Name), State: form.State}
	if product.Name == "" || utf8.RuneCountInString(product.Name) > 100 {
		form.Errors["name"] = "Укажи название длиной от 1 до 100 символов."
	}
	switch product.State {
	case "", "сухой", "сырой", "готовый":
	default:
		form.Errors["state"] = "Выбери состояние из списка."
	}
	fields := []*float64{&product.Per100.Protein, &product.Per100.Fat, &product.Per100.Carbs}
	missingNumber := false
	for index := range form.Numbers {
		field := &form.Numbers[index]
		field.Value = r.PostForm.Get(field.Key)
		value, ok := decimal(field.Value)
		if !ok {
			field.Error = "Введи число от 0 до 1 000 000."
			missingNumber = true
		}
		*fields[index] = value
	}
	if !missingNumber {
		form.Energy = calories(product.Per100.Kcal())
	}
	return product, form, len(form.Errors) == 0 && !missingNumber
}

func (a app) deleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteProduct(r.Context(), pathID(r)); err != nil {
		a.fail(w, r, err)
		return
	}
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
