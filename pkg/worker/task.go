package worker

import "context"

type TaskFn func(context.Context)

type task struct {
	fn TaskFn
	id int
}
