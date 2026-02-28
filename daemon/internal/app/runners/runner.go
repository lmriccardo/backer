package runners

import (
	"context"
	"log"
	"sync"

	"github.com/lmriccardo/backer/deamon/internal/domain"
)

type BackerRunner struct {
	ch  <-chan *domain.Job // Read-only channel for getting jobs to run
	ctx context.Context    // Shutdown context
	wg  *sync.WaitGroup    // Runner wait group from the engine
}

func NewBackerRunner(
	ctx context.Context, ch <-chan *domain.Job, wg *sync.WaitGroup,
) *BackerRunner {
	runner := &BackerRunner{ch: ch, ctx: ctx, wg: wg}
	return runner
}

func (b *BackerRunner) exec(job *domain.Job) {
	log.Printf("Running job: %s\n", job.Name)
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

			b.exec(job)
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

func NewRunningEngine(ctx context.Context, nWorkers int, ch <-chan *domain.Job) *RunningEngine {
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
