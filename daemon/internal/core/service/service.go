package service

import (
	"context"

	"github.com/lmriccardo/backer/deamon/internal/utils"
	"github.com/lmriccardo/backer/deamon/internal/utils/ilock"
)

type Service struct {
	IRegistry // Embed the actual registry

	HomeDir string // The home folder of backerd
	LogDir  string // Default log folder

	ctx  context.Context // Interrupt context
	lock *ilock.InstanceLock
}

func NewService(ctx context.Context) (*Service, error) {
	var err error

	instance_lock, err := ilock.NewInstanceLock(ilock.LOCK_NAME)
	if err != nil {
		return nil, err
	}

	service := &Service{ctx: ctx, lock: instance_lock}

	// Initialize the service with paths
	if service.HomeDir, err = utils.BackerHome(); err != nil {
		return nil, err
	}

	if service.LogDir, err = utils.BackerLogHome(); err != nil {
		return nil, err
	}

	// Initialize the registry
	if service.IRegistry, err = NewRegistry(ctx); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) Close() {
	s.IRegistry.Close() // Close the registry
	s.lock.Close()      // Release the lock
}
