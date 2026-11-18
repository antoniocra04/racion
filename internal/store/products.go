package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/antoniocra04/racion/internal/ration"
)

func (s *Store) Products(ctx context.Context) ([]ration.Product, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, revision, name, state, kcal, protein, fat, carbs FROM products ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []ration.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (s *Store) Product(ctx context.Context, id int64) (ration.Product, error) {
	product, err := scanProduct(s.db.QueryRowContext(ctx,
		"SELECT id, revision, name, state, kcal, protein, fat, carbs FROM products WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return product, ErrNotFound
	}
	return product, err
}

func scanProduct(row interface{ Scan(...any) error }) (ration.Product, error) {
	var p ration.Product
	err := row.Scan(&p.ID, &p.Revision, &p.Name, &p.State, &p.Per100.Kcal, &p.Per100.Protein, &p.Per100.Fat, &p.Per100.Carbs)
	return p, err
}

func (s *Store) SaveProduct(ctx context.Context, p ration.Product, creationToken string) (int64, error) {
	values := []any{p.Name, p.State, p.Per100.Kcal, p.Per100.Protein, p.Per100.Fat, p.Per100.Carbs}
	if p.ID == 0 {
		var id int64
		err := s.db.QueryRowContext(ctx, `
			INSERT INTO products(name, state, kcal, protein, fat, carbs, creation_token)
			VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''))
			ON CONFLICT(creation_token) DO UPDATE SET creation_token = excluded.creation_token
			RETURNING id`, append(values, creationToken)...).Scan(&id)
		return id, err
	}
	err := changed(s.db.ExecContext(ctx, `
		UPDATE products
		SET name = ?, state = ?, kcal = ?, protein = ?, fat = ?, carbs = ?, revision = revision + 1
		WHERE id = ? AND revision = ?`, append(values, p.ID, p.Revision)...))
	if errors.Is(err, ErrNotFound) {
		return 0, ErrProductConflict
	}
	return p.ID, err
}

func (s *Store) DeleteProduct(ctx context.Context, id int64) error {
	err := changed(s.db.ExecContext(ctx, `
		DELETE FROM products
		WHERE id = ? AND NOT EXISTS (SELECT 1 FROM meal_items WHERE product_id = products.id)`, id))
	if errors.Is(err, ErrNotFound) {
		if _, lookupErr := s.Product(ctx, id); lookupErr != nil {
			return lookupErr
		}
		return ErrProductInUse
	}
	return err
}
