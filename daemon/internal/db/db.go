package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

func AcquireInstanceLock(ctx context.Context, db *sql.DB) error {
	// WAL + single-writer is fine, but this prevents "two schedulers" bugs.
	// This lock is held as long as the connection remains open.
	_, err := db.ExecContext(ctx, `BEGIN IMMEDIATE;`)
	if err != nil {
		return fmt.Errorf("cannot acquire instance lock (another backerd running?): %w", err)
	}
	return nil
}

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
		`PRAGMA busy_timeout = 5000;`,  // ms: wait for locks instead of failing fast
		`PRAGMA synchronous = NORMAL;`, // good durability/perf tradeoff with WAL
	}

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("apply pragma %q: %w", p, err)
		}
	}

	return db, nil
}
