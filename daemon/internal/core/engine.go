package core

import (
	"context"
	"log"

	"github.com/lmriccardo/backer/deamon/internal/core/domain"
	"github.com/lmriccardo/backer/deamon/internal/core/runners"
	"github.com/lmriccardo/backer/deamon/internal/core/service"
)

type ServiceEngine struct {
	Service *service.Service // The actual service with the registry

	rEngine *runners.RunningEngine // The complete running engine
	runCh   chan *domain.JobRun    // Job runs channel
	ctx     context.Context        // Context for closure
}

func NewServiceEngine(ctx context.Context, nWorkers int) *ServiceEngine {
	// Intialize the service engine
	sEngine := &ServiceEngine{}

	// First create the service and the job db registry
	service, err := service.NewService(ctx, nWorkers)
	if err != nil {
		log.Fatalf("when creating the service: %v", err)
	}

	// Create all the runners with the running engine
	rChannel := service.IRegistry.GetTaskChannel()
	rEngine := runners.NewRunningEngine(ctx, nWorkers, rChannel)

	// Setup runs communication between the runner engine and the service
	runChannel := make(chan *domain.JobRun, 100)
	service.IRegistry.SetRunChannel(runChannel)
	rEngine.SetRunChannel(runChannel)

	sEngine.Service = service
	sEngine.rEngine = rEngine
	sEngine.runCh = runChannel
	sEngine.ctx = ctx

	// Finally start the engine
	sEngine.rEngine.Run()

	return sEngine
}

func (s *ServiceEngine) Run() {
	s.rEngine.Run()
}

func (s *ServiceEngine) Close() {
	close(s.runCh) // First close the runs channel

	s.Service.Close() // Cleanup the service
	s.rEngine.Stop()  // Stop the runner engine
}
