package ration

const (
	ProteinKcalPerGram = 4
	FatKcalPerGram     = 9
	CarbsKcalPerGram   = 4
)

// Nutrition describes the macronutrients of a quantity of food.
type Nutrition struct {
	Protein float64
	Fat     float64
	Carbs   float64
}

func (n Nutrition) Kcal() float64 {
	return n.Protein*ProteinKcalPerGram + n.Fat*FatKcalPerGram + n.Carbs*CarbsKcalPerGram
}

func (n Nutrition) scale(factor float64) Nutrition {
	return Nutrition{n.Protein * factor, n.Fat * factor, n.Carbs * factor}
}

func (n Nutrition) add(other Nutrition) Nutrition {
	return Nutrition{n.Protein + other.Protein, n.Fat + other.Fat, n.Carbs + other.Carbs}
}

type Product struct {
	ID       int64
	Revision int64
	Name     string
	State    string
	Per100   Nutrition
}

type Item struct {
	Product Product
	Grams   float64
}

func (i Item) Nutrition() Nutrition {
	return i.Product.Per100.scale(i.Grams / 100)
}

type Meal struct {
	Name  string
	Items []Item
}

func (m Meal) Nutrition() Nutrition {
	var total Nutrition
	for _, item := range m.Items {
		total = total.add(item.Nutrition())
	}
	return total
}

type Plan struct {
	ID       int64
	Revision int64
	Name     string
	Active   bool
	Meals    []Meal
}

func (p Plan) Nutrition() Nutrition {
	var total Nutrition
	for _, meal := range p.Meals {
		total = total.add(meal.Nutrition())
	}
	return total
}
