package migrations

import (
	"context"
	"database/sql"
)

func MigrationV2(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE runs (
			id TEXT NOT NULL UNIQUE PRIMARY KEY,
			job_id TEXT NOT NULL,
			started_at TEXT NOT NULL,
			one_shot INTEGER NOT NULL,
			finished_at TEXT,
			status TEXT NOT NULL,          -- e.g. "success", "failure", "running"
			exit_code INTEGER,
			error TEXT,
			log_path TEXT,

			FOREIGN KEY(job_id) REFERENCES jobs(name) ON DELETE CASCADE
		);

		CREATE INDEX idx_runs_job_id_started ON runs(job_id, started_at);
	`)
	return err
}
