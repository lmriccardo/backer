package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lmriccardo/backer/deamon/internal/core/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
	"github.com/stretchr/testify/require"
)

var _service *service.Service

func mustNewService(ctx context.Context) *service.Service {
	new_service, err := service.NewService(ctx, 0)
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
		backer_dir, _ := utils.BackerHome()
		os.RemoveAll(backer_dir)
		_service.Close()
		cancel()
		os.Exit(exit_code)
	}()

	_service = mustNewService(ctx)
	exit_code = m.Run()
}

// Tests that multiple service cannot be created given the lock
func TestMultipleService(t *testing.T) {
	_, err := service.NewService(context.Background(), 0)
	require.NotNil(t, err, "creating multiple service must raise an error")
}
