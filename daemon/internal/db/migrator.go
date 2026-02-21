package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/lmriccardo/backer/deamon/internal/db/migrations"
	"github.com/lmriccardo/backer/deamon/internal/platform/version"
)

type MigrationFunc func(context.Context, *sql.Tx) error

var migration_table = map[int]MigrationFunc{
	1: migrations.MigrationV1, 2: migrations.MigrationV2,
}

// EnsureSchema Ensures that the current registry DB is at the
// latest version, applying migrations if required
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	// 1. Create and validate the schema_version table
	if err := createSchemaVersion(ctx, db); err != nil {
		return err
	}

	// 2. Query the current version of the schema
	curr_version, err := currentVersion(ctx, db)
	if err != nil {
		return err
	}

	// 3. Update the current schema version up to latest one
	if curr_version < version.LATEST_DB_VERSION {
		log.Printf(
			"Current schema version %d is too old ( Latest is %d )",
			curr_version,
			version.LATEST_DB_VERSION,
		)
	}

	for curr_version < version.LATEST_DB_VERSION {
		next_version := curr_version + 1
		if err := applyMigration(ctx, db, next_version); err != nil {
			return err
		}
		curr_version = next_version
	}

	return nil
}

// createSchemaVersion creates the schema_version table, ensures that
// there is only one row ( if the table already exists ), and insert
// a new default entry with value 0 if the table was not existing
func createSchemaVersion(ctx context.Context, db *sql.DB) error {
	// Create the schema_version table if it does not exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	// Ensure that there is only one row in the table
	var rows_count int
	const count_query = `SELECT COUNT(*) FROM schema_version;`
	err = db.QueryRowContext(ctx, count_query).Scan(&rows_count)
	if err != nil {
		return fmt.Errorf("count schema_version: %w", err)
	}

	// If there is no row, insert the 0-default one
	if rows_count == 0 {
		log.Println("schema_version table not found; Creating default entry")
		_, err = db.ExecContext(ctx, `INSERT INTO schema_version(version) VALUES (0);`)
		if err != nil {
			return fmt.Errorf("init schema_version row: %w", err)
		}
	}

	return nil
}

// currentVersion Returns the current version from the schema_version table
func currentVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	const version_query = `SELECT version FROM schema_version LIMIT 1;`
	err := db.QueryRowContext(ctx, version_query).Scan(&version)
	if err != nil {
		return -1, fmt.Errorf("read schema_version: %w", err)
	}
	return version, nil
}

// applyMigration Applies the input migration version
func applyMigration(ctx context.Context, db *sql.DB, version int) error {
	err := WithTx(db, ctx, func(tx *sql.Tx) error {
		// Apply the migration stored in the table
		migration_fn, ok := migration_table[version]
		if !ok {
			return fmt.Errorf("unknown migration version: %d", version)
		}

		log.Printf("Performing migration V%d", version)
		if err := migration_fn(ctx, tx); err != nil {
			return fmt.Errorf("migrate v%d: %w", version, err)
		}

		// Finally update the schema version
		const update_query = `UPDATE schema_version SET version = ?;`
		_, err := tx.ExecContext(ctx, update_query, version)
		if err != nil {
			return fmt.Errorf("update schema_version=%d: %w", version, err)
		}
		return nil
	})
	return err
}
