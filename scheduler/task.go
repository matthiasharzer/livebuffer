package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/matthiasharzer/service-kit/logging"
)

var manager *taskManager

type WorkFunc = func() error

type task struct {
	work  WorkFunc
	runAt time.Time
}

func init() {
	ctx, cancel := context.WithCancel(context.Background())
	manager = &taskManager{
		mu:              sync.RWMutex{},
		shutdownCtx:     ctx,
		cancelCtx:       cancel,
		unfinishedTasks: make(map[*task]struct{}),
	}
}

type taskManager struct {
	mu              sync.RWMutex
	shutdownCtx     context.Context
	cancelCtx       context.CancelFunc
	unfinishedTasks map[*task]struct{}
}

func (m *taskManager) done() <-chan struct{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shutdownCtx.Done()
}

func (m *taskManager) runTask(task *task) {
	err := task.work()
	if err != nil {
		logging.Warn("scheduled task failed", "error", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.unfinishedTasks, task)
}

func (m *taskManager) awaitTask(task *task) {
	in := time.Until(task.runAt)
	for {
		select {
		case <-m.done():
			return
		case <-time.After(in):
			m.runTask(task)
			return
		}
	}
}

func (m *taskManager) scheduleTask(task *task) {
	m.mu.Lock()
	m.unfinishedTasks[task] = struct{}{}
	m.mu.Unlock()
	go m.awaitTask(task)
}

func (m *taskManager) shutdown() {
	m.cancelCtx()
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.unfinishedTasks) > 0 {
		logging.Warn("shutting down with unfinished tasks", "count", len(m.unfinishedTasks))
	}
	m.unfinishedTasks = make(map[*task]struct{})
}

type TimeContext struct {
	runAt time.Time
}

func (t TimeContext) Run(work WorkFunc) {
	manager.scheduleTask(&task{
		work:  work,
		runAt: t.runAt,
	})
}

func In(in time.Duration) TimeContext {
	return TimeContext{
		runAt: time.Now().Add(in),
	}
}

func At(runAt time.Time) TimeContext {
	return TimeContext{
		runAt: runAt,
	}
}

func Shutdown() {
	manager.shutdown()
}
