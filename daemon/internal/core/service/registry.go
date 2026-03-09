package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	apirequests "github.com/lmriccardo/backer/deamon/internal/api/v1/requests"
	"github.com/lmriccardo/backer/deamon/internal/core/domain"
	"github.com/lmriccardo/backer/deamon/internal/db"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
	_ "modernc.org/sqlite"
)

type StatementsMap map[RegistryStatementType]*sql.Stmt
type RegistryStatementType int

const (
	ListAllJobs RegistryStatementType = iota
	ListJobsWithStatus
	SearchJobByName
	InsertNewJob
)

var REGISTRY_STATEMENTS = map[RegistryStatementType]string{
	ListAllJobs:        `SELECT id, name, enabled, config_json FROM jobs`,
	ListJobsWithStatus: `SELECT id, name, enabled, config_json FROM jobs WHERE enabled = ?`,
	SearchJobByName:    `SELECT EXISTS ( SELECT 1 FROM jobs WHERE name = ? )`,
	InsertNewJob:       `INSERT INTO jobs(name, enabled, config_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
}

const NOF_WORKERS_DEFAULT = 4

type IRegistry interface {
	// ListJobs list all jobs corresponding to the input status
	ListJobs(ctx context.Context, status domain.JobStatus, tx *sql.Tx) ([]domain.Job, error)
	CreateJob(ctx context.Context, job *apirequests.CreateJobRequest, tx *sql.Tx) error
	GetJob(ctx context.Context, name string, tx *sql.Tx) (*domain.Job, error)
	SearchJobByName(ctx context.Context, name string, tx *sql.Tx) (bool, error)
	RunJob(ctx context.Context, req *apirequests.RunJobRequest, tx *sql.Tx) (*domain.Run, error)

	// Clean up the entire registry
	Close()
	GetTaskChannel() <-chan *domain.JobRun
	SetRunChannel(ch <-chan *domain.JobRun)
}

type Registry struct {
	*TickeringRoaster

	db         *sql.DB                   // The registry database
	path       string                    // The path to the file database
	ctx        context.Context           // Interrupt context
	statements StatementsMap             // Map of prepared statements
	runCh      <-chan *domain.JobRun     // Channel for reading job runs
	pendRuns   map[string]*domain.JobRun // Maps with one shot pending runs (not stored into the roaster)
}

func NewRegistry(ctx context.Context, nWorkers int) (*Registry, error) {
	r := &Registry{ctx: ctx, statements: StatementsMap{}}

	// Get the path of the registry file
	var err error
	if r.path, err = utils.RegistryFile(); err != nil {
		return nil, err
	}

	// Initialize the registry
	if err := r.initRegistry(nWorkers); err != nil {
		return nil, err
	}

	return r, nil
}

func NewInMemRegistry(ctx context.Context, nWorkers int) (*Registry, error) {
	r := &Registry{
		path:       "file::memory:?cache=shared&_fk=1",
		ctx:        ctx,
		statements: StatementsMap{},
	}

	// Initialize the registry
	if err := r.initRegistry(nWorkers); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Registry) initRegistry(nWorkers int) error {
	// Initialize the task roaster embedded structure
	r.TickeringRoaster = nil
	if nWorkers > 0 {
		r.TickeringRoaster = NewTickeringRoaster(r.ctx, nWorkers)
	}

	// Initialize the registry
	if err := r.initDb(); err != nil {
		return err
	}

	// Load all the jobs from the registry
	if err := r.loadJobs(); err != nil {
		return err
	}

	// Starts the goroutine for handling completed runs pushed
	// into the channel by all runners
	go r.handleCompletedRuns()

	return nil
}

func (r *Registry) loadJobs() error {
	// Loads the jobs from the registry DB
	jobs, err := r.ListJobs(context.Background(), domain.JobStatusAll, nil)
	if err != nil {
		return err
	}

	// Put each loaded job into the roaster and the mapping
	if r.TickeringRoaster != nil {
		for _, job := range jobs {
			r.Push(&job)
		}
	}

	return nil
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

func (r *Registry) bindStatementToTx(ctx context.Context, tx *sql.Tx, t RegistryStatementType) *sql.Stmt {
	if tx != nil {
		return tx.StmtContext(ctx, r.statements[t])
	}
	return r.statements[t]
}

func (r *Registry) handleCompletedRuns() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case run, ok := <-r.runCh:
			if !ok {
				return
			}

			log.Printf("Completed run with Id %s", run.Id)

			// Save the run into the database
		}
	}
}

// WithTx provides a transaction wrapper. Every operation perform by the input
// function is than rolled back to the previous state. Returns an error if any.
func (r *Registry) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	return db.WithTx(r.db, ctx, fn)
}

func (r *Registry) ListJobs(ctx context.Context, status domain.JobStatus, tx *sql.Tx) ([]domain.Job, error) {
	var rows *sql.Rows // Query exec destination
	var err error      // Error raised by the query

	if status == domain.JobStatusAll {
		rows, err = r.bindStatementToTx(ctx, tx, ListAllJobs).QueryContext(ctx)
	} else {
		rows, err = r.bindStatementToTx(ctx, tx, ListJobsWithStatus).QueryContext(ctx, int(status))
	}

	// If any error during query, returns empty list and the error
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := []domain.Job{} // Prepare the result
	var job_config string
	for rows.Next() {
		// Scan the current row and adds the values into the job,
		// except for the job configuration which should be
		// decoded from the JSON string stored into the database
		currjob := domain.Job{}
		err = rows.Scan(&currjob.Id, &currjob.Name, &currjob.Status, &job_config)
		if err != nil {
			return nil, err
		}

		// Decode the JSON configuration from the string
		if err := utils.ToJsonWithObj(&currjob.Config, job_config); err != nil {
			return nil, err
		}

		jobs = append(jobs, currjob)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetJob returns the job associated with the unique input name, if exists,
// otherwise it will returns nil and an error is set.
func (r *Registry) GetJob(ctx context.Context, name string, tx *sql.Tx) (*domain.Job, error) {
	return nil, nil
}

// SearchJobByName returns (True, nil) if there exists a registered job with
// given name, otherwise it will returns (nil, InvalidJobNameError).
func (r *Registry) SearchJobByName(ctx context.Context, name string, tx *sql.Tx) (bool, error) {
	var exists bool
	row := r.bindStatementToTx(ctx, tx, SearchJobByName).QueryRowContext(ctx, name)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	current_error := (error)(nil)
	if !exists {
		current_error = NewInvalidJobNameError(name)
	}
	return exists, current_error
}

// CreateJob creates a job from the HTTP job request, converts the job
// description into a job configuration and finally save the job
// into the registry database
func (r *Registry) CreateJob(ctx context.Context, job *apirequests.CreateJobRequest, tx *sql.Tx) error {
	log.Printf("Registering new backup job with name: %s", job.Name)

	// 1. Before starting validating the input job, we should check that
	// there not exists another job with the same name. Job names are unique.
	result, err := r.SearchJobByName(ctx, job.Name, nil)
	_, ok := err.(*InvalidJobNameError)
	if err != nil && !ok {
		log.Printf("(SearchJobByName Error): %v", err.Error())
		return err
	}

	// If the job exists returns a new error
	if result {
		log.Printf("(Error): Job %v already registered", job.Name)
		return NewDuplicateJobNameError(job.Name)
	}

	// 2. Create the job structure from the request
	registry_job, err := createJob(job)
	if err != nil {
		log.Printf("(Error) when creating Job: %v", err.Error())
		return err
	}

	now := time.Now().UTC()
	formatted := now.Format("2006-01-02 15:04:05")

	// 3. Inser the job into the table
	stmt := r.bindStatementToTx(ctx, tx, InsertNewJob)
	config := utils.JSONToString(&registry_job.Config)
	_, err = stmt.ExecContext(ctx, registry_job.Name, registry_job.Status,
		config, formatted, formatted)

	if err != nil {
		log.Printf("(Database Execution Error): %v", err.Error())
		return NewDatabaseError(
			fmt.Sprintf("unable to insert new job %v", job.Name),
		)
	}

	// 4. Pushes the job into the task roaster for execution
	if r.TickeringRoaster != nil {
		r.Push(registry_job)
	}

	return nil
}

func (r *Registry) RunJob(ctx context.Context, req *apirequests.RunJobRequest, tx *sql.Tx) (*domain.Run, error) {
	// First we need to check that the job actually exists
	_, ok := r.GetTaskIndex(req.Name)
	if !ok {
		return nil, NewInvalidJobNameError(req.Name)
	}

	// If the job exists than we can create the run task and push into the channel
	job, _ := r.GetTask(req.Name)
	job.Name = fmt.Sprintf("%s_ONESHOT", job.Name)
	task := r.CreateJobRunTask(job, req.DryRun, true)
	r.ready <- task

	return nil, nil
}

func (r *Registry) Close() {
	// Close all the open statements and the database
	for _, v := range r.statements {
		v.Close()
	}
	r.db.Close()

	// Close the write channel for jobs
	r.TickeringRoaster.Close()
}

func (r *Registry) GetTaskChannel() <-chan *domain.JobRun {
	return r.TickeringRoaster.RChannel()
}

func (r *Registry) SetRunChannel(ch <-chan *domain.JobRun) {
	r.runCh = ch
}
