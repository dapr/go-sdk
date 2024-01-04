package workflow

import (
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/types/known/wrapperspb"

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

type callActivityOption func(*callActivityOptions) error

type callActivityOptions struct {
	rawInput *wrapperspb.StringValue
}

func ActivityInput(input any) callActivityOption {
	return func(opt *callActivityOptions) error {
		data, err := marshalData(input)
		if err != nil {
			return err
		}
		opt.rawInput = wrapperspb.String(string(data))
		return nil
	}
}

func marshalData(input any) ([]byte, error) {
	if input == nil {
		return nil, errors.New("empty input")
	}
	return json.Marshal(input)
}
