package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/antoniocra04/racion/internal/ration"
)

func (s *Store) Plans(ctx context.Context) ([]ration.Plan, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT plans.id, plans.name, plans.revision, COALESCE(plans.id = settings.active_plan_id, 0),
		       COALESCE(meals.id, 0), COALESCE(meals.name, ''),
		       COALESCE(products.id, 0), COALESCE(products.name, ''), COALESCE(products.state, ''),
		       COALESCE(products.kcal, 0), COALESCE(products.protein, 0), COALESCE(products.fat, 0),
		       COALESCE(products.carbs, 0), COALESCE(meal_items.grams, 0)
		FROM plans CROSS JOIN settings
		LEFT JOIN meals ON meals.plan_id = plans.id
		LEFT JOIN meal_items ON meal_items.meal_id = meals.id
		LEFT JOIN products ON products.id = meal_items.product_id
		ORDER BY plans.id, meals.position, meal_items.position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []ration.Plan
	var currentMeal int64
	for rows.Next() {
		var plan ration.Plan
		var mealID int64
		var meal ration.Meal
		var item ration.Item
		p := &item.Product
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.Revision, &plan.Active,
			&mealID, &meal.Name, &p.ID, &p.Name, &p.State, &p.Per100.Kcal, &p.Per100.Protein,
			&p.Per100.Fat, &p.Per100.Carbs, &item.Grams); err != nil {
			return nil, err
		}
		if len(plans) == 0 || plans[len(plans)-1].ID != plan.ID {
			plans = append(plans, plan)
			currentMeal = 0
		}
		if mealID == 0 {
			continue
		}
		current := &plans[len(plans)-1]
		if currentMeal != mealID {
			current.Meals = append(current.Meals, meal)
			currentMeal = mealID
		}
		if p.ID != 0 {
			lastMeal := &current.Meals[len(current.Meals)-1]
			lastMeal.Items = append(lastMeal.Items, item)
		}
	}
	return plans, rows.Err()
}

func (s *Store) SavePlan(ctx context.Context, plan ration.Plan) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	id := plan.ID
	if id == 0 {
		result, err := tx.ExecContext(ctx, "INSERT INTO plans(name) VALUES (?)", plan.Name)
		if err != nil {
			return 0, err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}
	} else {
		err := changed(tx.ExecContext(ctx, `
			UPDATE plans SET name = ?, revision = revision + 1 WHERE id = ? AND revision = ?`,
			plan.Name, id, plan.Revision))
		if errors.Is(err, ErrNotFound) {
			return 0, ErrConflict
		}
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM meals WHERE plan_id = ?", id); err != nil {
			return 0, err
		}
	}
	if err := insertMeals(ctx, tx, id, plan.Meals); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE settings SET active_plan_id = ? WHERE active_plan_id IS NULL", id); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func insertMeals(ctx context.Context, tx *sql.Tx, planID int64, meals []ration.Meal) error {
	for position, meal := range meals {
		inserted, err := tx.ExecContext(ctx, "INSERT INTO meals(plan_id, position, name) VALUES (?, ?, ?)",
			planID, position, meal.Name)
		if err != nil {
			return err
		}
		mealID, err := inserted.LastInsertId()
		if err != nil {
			return err
		}
		for itemPosition, item := range meal.Items {
			err := changed(tx.ExecContext(ctx, `
				INSERT INTO meal_items(meal_id, product_id, position, grams)
				SELECT ?, id, ?, ? FROM products WHERE id = ?`,
				mealID, itemPosition, item.Grams, item.Product.ID))
			if errors.Is(err, ErrNotFound) {
				return ErrProductMissing
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) ActivatePlan(ctx context.Context, id int64) error {
	return changed(s.db.ExecContext(ctx, `
		UPDATE settings SET active_plan_id = ? WHERE EXISTS (SELECT 1 FROM plans WHERE id = ?)`, id, id))
}

func (s *Store) DeletePlan(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var protected bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM plans WHERE id != ?)
		  AND EXISTS (SELECT 1 FROM settings WHERE active_plan_id = ?)`, id, id).Scan(&protected); err != nil {
		return err
	}
	if protected {
		return ErrActivePlan
	}
	if err := changed(tx.ExecContext(ctx, "DELETE FROM plans WHERE id = ?", id)); err != nil {
		return err
	}
	return tx.Commit()
}
