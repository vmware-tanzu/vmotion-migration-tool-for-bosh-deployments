package worker

import (
	"context"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
)

type Pool struct {
	taskCount   int
	workerCount int
	parentCtx   context.Context
	taskQueue   chan task
}

// NewPool will create a Pool of workers
func NewPool(workerCount int) *Pool {
	return &Pool{
		workerCount: workerCount,
		taskQueue:   make(chan task),
	}
}

func (p *Pool) Start(ctx context.Context) {
	p.parentCtx = ctx
	p.run()
}

func (p *Pool) AddTask(taskFn TaskFn) {
	p.taskCount++
	p.taskQueue <- task{
		fn: taskFn,
		id: p.taskCount,
	}
}

func (p *Pool) run() {
	l := log.WithoutContext()
	for i := 1; i < p.workerCount+1; i++ {
		l.Debugf("Created worker %d", i)

		go func(workerID int) {
			for task := range p.taskQueue {
				l.Debugf("Worker %d started processing task %d", workerID, task.id)
				ctx := context.WithValue(p.parentCtx, log.TaskIDKey, task.id)
				task.fn(ctx)
				l.Debugf("Worker %d finished processing task %d", workerID, task.id)
			}
		}(i)
	}
}
