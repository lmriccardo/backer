package runners

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lmriccardo/backer/deamon/internal/core/domain"
)

type BackerRunner struct {
	ch  <-chan *domain.JobRun // Read-only channel for getting jobs to run
	rch chan<- *domain.JobRun // Write-only channel for pushing jobs run
	ctx context.Context       // Shutdown context
	wg  *sync.WaitGroup       // Runner wait group from the engine
}

func NewBackerRunner(
	ctx context.Context, ch <-chan *domain.JobRun, wg *sync.WaitGroup,
) *BackerRunner {
	runner := &BackerRunner{ch: ch, ctx: ctx, wg: wg}
	return runner
}

func (b *BackerRunner) exec(job *domain.Job) {
	log.Printf("Running job: %s\n", job.Name)
	time.Sleep(1 * time.Second)
}

func (b *BackerRunner) Run() {
	defer b.wg.Done() // Signal goroutine completition
	for {
		select {
		case <-b.ctx.Done():
			return
		case job, ok := <-b.ch:
			// If the channel closes then we need to
			// return and stopping processing jobs
			if !ok {
				return
			}

			// Set its status to running and communicate to the registry
			// that the job is running
			job.Status = domain.RunStatusRunning
			b.rch <- job

			// Execute the task
			b.exec(job.Job)

			// Set its status to completed and push it back into the
			// communication channel with the registry.
			job.Status = domain.RunStatusCompleted
			b.rch <- job
		}
	}
}

// RunningEngine creates and manage running lifetime of all
// runners for backup jobs
type RunningEngine struct {
	runners []*BackerRunner    // All the runners
	ctx     context.Context    // Shutdown context for the engine
	cancel  context.CancelFunc // Cancellation function for runner context
	wg      sync.WaitGroup     // Waiting group for synchronizing runner exit
}

func NewRunningEngine(ctx context.Context, nWorkers int, ch <-chan *domain.JobRun) *RunningEngine {
	engine := &RunningEngine{runners: []*BackerRunner{}, ctx: ctx}

	// Create all the runners
	rCtx, rCancel := context.WithCancel(ctx)
	engine.cancel = rCancel
	for range nWorkers {
		runner := NewBackerRunner(rCtx, ch, &engine.wg)
		engine.runners = append(engine.runners, runner)
	}

	return engine
}

func (e *RunningEngine) SetRunChannel(ch chan<- *domain.JobRun) {
	for _, runner := range e.runners {
		runner.rch = ch
	}
}

func (e *RunningEngine) Run() {
	for _, runner := range e.runners {
		e.wg.Add(1)
		go runner.Run()
	}

	log.Printf("Spinned up %d runners\n", len(e.runners))
}

func (e *RunningEngine) Stop() {
	e.cancel()  // Cancel the context for all runner
	e.wg.Wait() // Wait for all runner to exit thei routine

	log.Println("Running Engine Stopped")
}
