package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TickeringTask interface {
	Duration() time.Duration // Returns the duration of the ticker
}

type RoasterTask[T TickeringTask] struct {
	task   *T
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
}

type TickeringRoaster[T TickeringTask] struct {
	tasks []*RoasterTask[T] // The actual task list
	ready chan *T           // Indexes of free tasks
	mu    sync.RWMutex      // The mutex for concurrency
	ctx   context.Context   // The stopping context
}

func NewTickeringRoaster[T TickeringTask](ctx context.Context, nWorkers int) *TickeringRoaster[T] {
	readyChan := make(chan *T, nWorkers)
	roaster := &TickeringRoaster[T]{ready: readyChan, ctx: ctx}
	return roaster
}

func (a *TickeringRoaster[T]) tickTask(t *RoasterTask[T]) {
	defer t.ticker.Stop()
	for {
		select {
		case <-t.ticker.C:
			// Signal that this specific task is ready
			a.ready <- t.task
		case <-a.ctx.Done():
			return
		case <-t.ctx.Done():
			return
		}
	}
}

// Push pushes an element at the end of the roaster
func (a *TickeringRoaster[T]) Push(elem *T) {
	a.mu.Lock()
	defer a.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	taskWrapper := &RoasterTask[T]{
		task:   elem,
		ticker: time.NewTicker((*elem).Duration()),
		ctx:    ctx,
		cancel: cancel,
	}
	a.tasks = append(a.tasks, taskWrapper)

	// Starts a go routine for each task getting its own pulse
	go a.tickTask(taskWrapper)
}

func (a *TickeringRoaster[T]) Remove(idx int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if idx < 0 || idx >= len(a.tasks) {
		return fmt.Errorf(
			"out of range on current size %d: given %d",
			len(a.tasks),
			idx,
		)
	}

	task := a.tasks[idx]
	task.cancel()

	// Remove unordered (O(1)). we do not care about ordering
	// in this roaster of tasks.
	last := len(a.tasks) - 1
	a.tasks[idx] = a.tasks[last]
	a.tasks[last] = nil // avoid memory retention
	a.tasks = a.tasks[:last]

	return nil
}

func (a *TickeringRoaster[T]) Size() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.tasks)
}

// RChannel returns a read-only channel of tasks
func (a *TickeringRoaster[T]) RChannel() <-chan *T {
	return a.ready
}
