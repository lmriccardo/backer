package core

import (
	"context"
	"log"
)

type Service struct {
	HomeDir  string    // The home folder of backerd
	LogDir   string    // Default log folder
	Registry IRegistry // Pointer to a generic registry interface

	ctrlcCtx context.Context // Interrupt context
}

func NewService(ctx context.Context) *Service {
	service := &Service{ctrlcCtx: ctx}

	// Initialize the service with paths
	var err error
	if service.HomeDir, err = BackerHome(); err != nil {
		log.Fatal(err)
	}
	if service.LogDir, err = BackerLogHome(); err != nil {
		log.Fatal(err)
	}

	// Initialize the registry
	if service.Registry, err = NewRegistry(ctx); err != nil {
		log.Fatal(err)
	}

	return service
}
