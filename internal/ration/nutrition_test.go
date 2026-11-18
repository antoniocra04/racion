package ration

import "testing"

func TestMealAndDayTotals(t *testing.T) {
	food := Product{Per100: Nutrition{10, 4, 11}}
	portion := Item{Product: food, Grams: 150}
	if got := portion.Nutrition(); got != (Nutrition{15, 6, 16.5}) || got.Kcal() != 180 {
		t.Fatalf("portion nutrition = %+v", got)
	}
	plan := Plan{Meals: []Meal{
		{Items: []Item{portion, {Product: food, Grams: 50}}},
		{Items: []Item{{Product: food, Grams: 25}}},
	}}
	if got := plan.Meals[0].Nutrition(); got != (Nutrition{20, 8, 22}) || got.Kcal() != 240 {
		t.Fatalf("meal nutrition = %+v", got)
	}
	if got := plan.Nutrition(); got != (Nutrition{22.5, 9, 24.75}) || got.Kcal() != 270 {
		t.Fatalf("day nutrition = %+v", got)
	}
}
