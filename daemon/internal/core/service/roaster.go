package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lmriccardo/backer/deamon/internal/core/domain"
)

type idProvider struct {
	mu     sync.Mutex
	ids    map[string]int
	prefix string
}

func newIdProvider(prefix string) *idProvider {
	return &idProvider{ids: make(map[string]int), prefix: prefix}
}

func (i *idProvider) getId(name string) string {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Check if the target is present, if not register it
	curr_id, ok := i.ids[name]
	if !ok {
		i.ids[name] = 0
		curr_id = i.ids[name]
	}

	i.ids[name]++
	return fmt.Sprintf("%s%s_%d", i.prefix, name, curr_id)
}

type RoasterTask struct {
	task   *domain.JobRun
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
}

type TickeringRoaster struct {
	tasks       []*RoasterTask      // The actual task list
	ready       chan *domain.JobRun // Indexes of free tasks
	taskToIndex map[string]int      // Maps tasks name to index
	mu          sync.RWMutex        // The mutex for concurrency
	ctx         context.Context     // The stopping context
	idP         *idProvider         // Provides id generation for runs
}

func NewTickeringRoaster(ctx context.Context, nWorkers int) *TickeringRoaster {
	readyChan := make(chan *domain.JobRun, nWorkers)
	idP := newIdProvider("RUN_")
	roaster := &TickeringRoaster{
		ready:       readyChan,
		taskToIndex: map[string]int{},
		ctx:         ctx,
		idP:         idP,
	}

	return roaster
}

func (a *TickeringRoaster) CreateJobRunTask(
	job *domain.Job, dryRun bool, oneShot bool,
) *domain.JobRun {
	return &domain.JobRun{
		Job: job,
		Run: domain.Run{
			JobName: job.Name,
			Id:      a.idP.getId(job.Name),
			DryRun:  dryRun,
			Status:  domain.RunStatusWaiting,
			OneShot: oneShot,
		},
	}
}

func (a *TickeringRoaster) tickTask(t *RoasterTask) {
	defer t.ticker.Stop()
	for {
		select {
		case <-t.ticker.C:
			// Signal that this specific task is ready
			// First we also need to increment its run counter
			t.task.Id = a.idP.getId(t.task.JobName)
			a.ready <- t.task
		case <-a.ctx.Done():
			return
		case <-t.ctx.Done():
			return
		}
	}
}

// Push pushes an element at the end of the roaster
func (a *TickeringRoaster) Push(elem *domain.Job) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if the task already exists in the task list
	if _, ok := a.taskToIndex[elem.Name]; ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	taskWrapper := &RoasterTask{
		task:   a.CreateJobRunTask(elem, false, false),
		ticker: time.NewTicker((*elem).Duration()),
		ctx:    ctx,
		cancel: cancel,
	}
	a.taskToIndex[elem.Name] = len(a.tasks)
	a.tasks = append(a.tasks, taskWrapper)

	// Starts a go routine for each task getting its own pulse
	go a.tickTask(taskWrapper)
}

func (a *TickeringRoaster) Remove(idx int) error {
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

	// Remove the task from the map
	delete(a.taskToIndex, task.task.JobName)

	// Remove unordered (O(1)). we do not care about ordering
	// in this roaster of tasks.
	last := len(a.tasks) - 1
	a.tasks[idx] = a.tasks[last]
	a.tasks[last] = nil // avoid memory retention
	a.tasks = a.tasks[:last]

	// Set the new task lying at index position in the map
	a.taskToIndex[a.tasks[idx].task.JobName] = idx

	return nil
}

func (a *TickeringRoaster) Size() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.tasks)
}

// RChannel returns a read-only channel of tasks
func (a *TickeringRoaster) RChannel() <-chan *domain.JobRun {
	return a.ready
}

// GetTaskIndex returns the index associated with the input job name
func (a *TickeringRoaster) GetTaskIndex(name string) (int, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	idx, ok := a.taskToIndex[name]
	if !ok {
		return -1, false
	}
	return idx, true
}

func (a *TickeringRoaster) GetTask(name string) (*domain.Job, bool) {
	taskIndex, ok := a.GetTaskIndex(name)
	if !ok {
		return nil, false
	}

	// Then we can initiate the lock and release it once
	// the function ends
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tasks[taskIndex].task.Job, true
}

// Close closes the ready channel for pushing task to runners
func (a *TickeringRoaster) Close() {
	close(a.ready)
}
