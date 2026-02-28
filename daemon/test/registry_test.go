package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "embed"

	"github.com/google/go-cmp/cmp"
	"github.com/lmriccardo/backer/deamon/internal/app/service"
	"github.com/lmriccardo/backer/deamon/internal/domain"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
)

/*
This tests the registry database functionality.
*/

type RegistryFunc func(context.Context, *service.Registry, *sql.Tx) error

//go:embed testdata/insert_job.sql
var insertJobs string

var reg *service.Registry = nil

// This function ensures that the registry is always created. If an
// error occur when the registry is created then it fails immediately
func mustNewRegistry(t *testing.T) *service.Registry {
	reg, err := service.NewInMemRegistry(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func testListAllJobs(ctx context.Context, reg *service.Registry, tx *sql.Tx) error {
	jobs, err := reg.ListJobs(ctx, domain.JobStatusAll, tx)
	if err != nil {
		return err
	}

	if len(jobs) != 2 {
		return fmt.Errorf("expected 2 jobs, found: %d", len(jobs))
	}

	return nil
}

func testListEnabledJobs(ctx context.Context, reg *service.Registry, tx *sql.Tx) error {
	jobs, err := reg.ListJobs(ctx, domain.JobStatusEnabled, tx)
	if err != nil {
		return err
	}

	if len(jobs) != 1 {
		return fmt.Errorf("expected 1 jobs, found: %d", len(jobs))
	}

	expected := domain.Job{
		Id:     1,
		Name:   "simple_backup",
		Status: domain.JobStatusEnabled,
		Config: domain.JobConfig{
			Name:        "simple_backup",
			Log:         domain.LogConfig{Path: "/path/to/log_folder"},
			Compression: false,
			Command:     []string{"ciao"},
			Notify:      nil,
		},
	}

	if diff := cmp.Diff(expected, jobs[0]); diff != "" {
		return fmt.Errorf("mismatch (-want +got):\n%s", diff)
	}

	return nil
}

func testListDisabledJobs(ctx context.Context, reg *service.Registry, tx *sql.Tx) error {
	jobs, err := reg.ListJobs(ctx, domain.JobStatusDisabled, tx)
	if err != nil {
		return err
	}

	if len(jobs) != 1 {
		return fmt.Errorf("expected 1 jobs, found: %d", len(jobs))
	}

	expected := domain.Job{
		Id:     2,
		Name:   "full_backup",
		Status: domain.JobStatusDisabled,
		Config: domain.JobConfig{
			Name:        "full_backup",
			Log:         domain.LogConfig{Path: "/path/to/log_folder1"},
			Compression: true,
			Command:     []string{"ls"},
			Notify:      nil,
		},
	}

	if diff := cmp.Diff(expected, jobs[0]); diff != "" {
		return fmt.Errorf("mismatch (-want +got):\n%s", diff)
	}

	return nil
}

// This tests if the registry DB can be easily created and all
// migrations can be applied without breaking anything
func TestMigrations(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Set the HOME environment variable to trick where the registry
	// db file should be created at.
	os.Setenv("HOME", filepath.Join(cwd, "testdata"))

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		// Remove the path created for the registry
		backer_dir, _ := utils.BackerHome()
		if err := os.RemoveAll(backer_dir); err != nil {
			t.Error(err)
		}
		cancel()
	}()

	if _, err := service.NewRegistry(ctx, 0); err != nil {
		t.Error(err)
	}
}

// This tests the list jobs service provided by the registry.
// In case JobStatusAll is given, the function must returns
// exactly 2 jobs, those being added by the fixture insertJobs.
func TestListJobs(t *testing.T) {
	ctx := t.Context()

	// Create the registry only if it does not exists.
	// In this way we can achieve persistance among tests.
	if reg == nil {
		reg = mustNewRegistry(t)
	}

	defer reg.Close()

	funcs := []RegistryFunc{testListAllJobs, testListEnabledJobs, testListDisabledJobs}

	err := reg.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, insertJobs); err != nil {
			return err
		}

		// Test List all jobs
		for _, fn := range funcs {
			if err := fn(ctx, reg, tx); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
