package workflow

import (
	"context"

	"github.com/microsoft/durabletask-go/task"
)

type ActivityContext struct {
	ctx task.ActivityContext
}

func (wfac *ActivityContext) GetInput(v interface{}) error {
	return wfac.ctx.GetInput(&v)
}

func (wfac *ActivityContext) Context() context.Context {
	return wfac.ctx.Context()
}
