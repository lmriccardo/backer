package migrations

import (
	"context"
	"database/sql"
)

// migrationV1 Introduces the JOBS table
func MigrationV1(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			enabled INTEGER NOT NULL CHECK (enabled IN (0,1)),
			config_json TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE INDEX idx_jobs_enabled ON jobs(enabled);
		CREATE INDEX idx_jobs_name ON jobs(name);
	`)
	return err
}
