package core

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/lmriccardo/backer/deamon/internal/db"
	"github.com/lmriccardo/backer/deamon/internal/db/migrations"
	_ "modernc.org/sqlite"
)

func loadRegistryDb(path string) (*sql.DB, error) {
	// 1. Open the database from the file path
	log.Printf("Loading Registry DB from %s\n", path)

	d, err := db.OpenDB(path)
	if err != nil {
		return nil, err
	}

	// 2. Apply migrations if necessary
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := migrations.EnsureSchema(ctx, d); err != nil {
		return nil, err
	}

	// We would like only one daemon instance. If there are two,
	// then one should fail immediately
	if err := db.AcquireInstanceLock(ctx, d); err != nil {
		_ = d.Close()
		return nil, err
	}

	return d, nil
}

type IRegistry interface {
	// ListJobs( ctx context.Context )
}

type Registry struct {
	path string          // The path to the file database
	db   *sql.DB         // The registry database
	ctx  context.Context // Interrupt context
}

func NewRegistry(ctx context.Context) (*Registry, error) {
	r := &Registry{ctx: ctx}

	// Initialize the registry
	if err := r.init(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Registry) init() error {
	var err error
	if r.path, err = RegistryFile(); err != nil {
		return err
	}

	// Load the registry database
	if r.db, err = loadRegistryDb(r.path); err != nil {
		return err
	}

	return nil
}
