package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Good sane defaults for daemons
	db.SetMaxOpenConns(1) // SQLite likes single-writer
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check context ping using the database
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	pragmas := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA synchronous = NORMAL;`,
	}

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("apply pragma %q: %w", p, err)
		}
	}

	return db, nil
}

func LoadFromPath(path string) (*sql.DB, error) {
	// 1. Open the database from the file path
	d, err := OpenDB(path)
	if err != nil {
		return nil, err
	}

	// 2. Apply migrations if necessary
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := EnsureSchema(ctx, d); err != nil {
		return nil, err
	}

	return d, nil
}

func WithTx(db *sql.DB, ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
