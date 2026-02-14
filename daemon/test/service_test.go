package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lmriccardo/backer/deamon/internal/core"
	"github.com/stretchr/testify/require"
)

var service *core.Service

func mustNewService(ctx context.Context) *core.Service {
	new_service, err := core.NewService(ctx)
	if err != nil {
		panic(err)
	}
	return new_service
}

func TestMain(m *testing.M) {
	// Set the HOME environment variable to trick where the registry
	// db file should be created at.
	cwd, _ := os.Getwd()
	os.Setenv("HOME", filepath.Join(cwd, "testdata"))

	ctx, cancel := context.WithCancel(context.Background())
	var exit_code int

	defer func() {
		// Remove the path created for the registry
		backer_dir, _ := core.BackerHome()
		os.RemoveAll(backer_dir)
		service.Close()
		cancel()
		os.Exit(exit_code)
	}()

	service = mustNewService(ctx)
	exit_code = m.Run()
}

// Tests that multiple service cannot be created given the lock
func TestMultipleService(t *testing.T) {
	_, err := core.NewService(context.Background())
	require.NotNil(t, err, "creating multiple service must raise an error")
}
