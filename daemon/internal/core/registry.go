package core

import (
	"context"
	"database/sql"
	"log"

	"github.com/lmriccardo/backer/deamon/internal/db"
	_ "modernc.org/sqlite"
)

type StatementsMap map[RegistryStatementType]*sql.Stmt
type RegistryStatementType int

const (
	ListAllJobs RegistryStatementType = iota
	ListJobsWithStatus
)

var REGISTRY_STATEMENTS = map[RegistryStatementType]string{
	ListAllJobs:        `SELECT id, name, enabled, config_json FROM jobs`,
	ListJobsWithStatus: `SELECT id, name, enabled, config_json FROM jobs WHERE enabled = ?`,
}

type IRegistry interface {
	// ListJobs list all jobs corresponding to the input status
	ListJobs(ctx context.Context, status JobStatus, tx *sql.Tx) ([]Job, error)

	// Clean up the entire registry
	Close()
}

type Registry struct {
	db         *sql.DB         // The registry database
	path       string          // The path to the file database
	ctx        context.Context // Interrupt context
	statements StatementsMap   // Map of prepared statements
}

func NewRegistry(ctx context.Context) (*Registry, error) {
	r := &Registry{ctx: ctx, statements: StatementsMap{}}

	// Get the path of the registry file
	var err error
	if r.path, err = RegistryFile(); err != nil {
		return nil, err
	}

	// Initialize the registry
	if err := r.initDb(); err != nil {
		return nil, err
	}

	return r, nil
}

func NewInMemRegistry(ctx context.Context) (*Registry, error) {
	r := &Registry{
		path:       "file::memory:?cache=shared&_fk=1",
		ctx:        ctx,
		statements: StatementsMap{},
	}

	// Initialize the registry
	if err := r.initDb(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Registry) prepareStatements() error {
	for k, v := range REGISTRY_STATEMENTS {
		stmt, err := r.db.PrepareContext(r.ctx, v)
		if err != nil {
			return err
		}

		r.statements[k] = stmt
	}

	return nil
}

func (r *Registry) initDb() error {
	var err error

	// Load the registry database
	log.Printf("Loading Registry DB from %s\n", r.path)
	if r.db, err = db.LoadFromPath(r.path); err != nil {
		return err
	}

	// Prepare all the statements for future usage
	if err := r.prepareStatements(); err != nil {
		return err
	}

	return nil
}

// WithTx provides a transaction wrapper. Every operation perform by the input
// function is than rolled back to the previous state. Returns an error if any.
func (r *Registry) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Registry) bindStatementToTx(ctx context.Context, tx *sql.Tx, t RegistryStatementType) *sql.Stmt {
	if tx != nil {
		return tx.StmtContext(ctx, r.statements[t])
	}
	return r.statements[t]
}

func (r *Registry) ListJobs(ctx context.Context, status JobStatus, tx *sql.Tx) ([]Job, error) {
	var rows *sql.Rows // Query exec destination
	var err error      // Error raised by the query

	if status == JobStatusAll {
		rows, err = r.bindStatementToTx(ctx, tx, ListAllJobs).QueryContext(ctx)
	} else {
		rows, err = r.bindStatementToTx(ctx, tx, ListJobsWithStatus).QueryContext(ctx, int(status))
	}

	// If any error during query, returns empty list and the error
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := []Job{} // Prepare the result
	var job_config string
	for rows.Next() {
		// Scan the current row and adds the values into the job,
		// except for the job configuration which should be
		// decoded from the JSON string stored into the database
		currjob := Job{}
		err = rows.Scan(&currjob.Id, &currjob.Name, &currjob.Status, &job_config)
		if err != nil {
			return nil, err
		}

		// Decode the JSON configuration from the string
		if err := ToJsonWithObj(&currjob.Config, job_config); err != nil {
			return nil, err
		}

		jobs = append(jobs, currjob)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *Registry) Close() {
	for _, v := range r.statements {
		v.Close()
	}
	r.db.Close()
}
