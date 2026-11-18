package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound        = errors.New("record not found")
	ErrProductInUse    = errors.New("product is used in a plan")
	ErrProductMissing  = errors.New("plan references a missing product")
	ErrProductConflict = errors.New("product has changed since it was opened")
	ErrActivePlan      = errors.New("select another active plan first")
	ErrConflict        = errors.New("plan has changed since it was opened")
)

//go:embed schema.sql
var schema string

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	dsn := url.URL{Scheme: "file", Path: path}
	dsn.RawQuery = url.Values{
		"_pragma": {"busy_timeout(5000)", "foreign_keys(1)"},
		"_txlock": {"immediate"},
	}.Encode()
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.migrate(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("open database: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var version int
	if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	switch version {
	case 0:
		if _, err := tx.ExecContext(ctx, schema); err != nil {
			return err
		}
	case 1:
	default:
		return fmt.Errorf("unsupported database version %d", version)
	}
	return tx.Commit()
}

// Backup writes a consistent snapshot to a new file without overwriting existing data.
func (s *Store) Backup(ctx context.Context, path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(path), ".ration-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	temp := filepath.Join(dir, "snapshot.db")
	if _, err := s.db.ExecContext(ctx, "VACUUM INTO ?", temp); err != nil {
		return err
	}
	if err := os.Chmod(temp, 0600); err != nil {
		return err
	}
	return os.Link(temp, path)
}

func changed(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return ErrNotFound
	}
	return err
}
