package web

import (
	"net/http"
	"strconv"

	"github.com/antoniocra04/racion/internal/ration"
	"github.com/antoniocra04/racion/internal/store"
)

func (a app) viewRation(w http.ResponseWriter, r *http.Request) {
	plans, err := a.store.Plans(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	data := page{Title: "Дневной рацион", Screen: "day", Plans: plans}
	if r.PathValue("id") != "" {
		data.Plan = findPlan(plans, pathID(r))
		if data.Plan == nil {
			a.fail(w, r, store.ErrNotFound)
			return
		}
	} else {
		for i := range plans {
			if plans[i].Active {
				data.Plan = &plans[i]
				break
			}
		}
	}
	a.render(w, r, http.StatusOK, data)
}

func findPlan(plans []ration.Plan, id int64) *ration.Plan {
	for i := range plans {
		if plans[i].ID == id {
			return &plans[i]
		}
	}
	return nil
}

func (a app) listPlans(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, http.StatusOK, page{Title: "Версии рациона", Screen: "plans"})
}

func (a app) activate(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r, 4<<10) {
		return
	}
	id, err := strconv.ParseInt(r.PostForm.Get("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "Выбери сохранённый рацион.", http.StatusBadRequest)
		return
	}
	if err := a.store.ActivatePlan(r.Context(), id); err != nil {
		a.fail(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a app) deletePlan(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeletePlan(r.Context(), pathID(r)); err != nil {
		a.fail(w, r, err)
		return
	}
	http.Redirect(w, r, "/plans", http.StatusSeeOther)
}
