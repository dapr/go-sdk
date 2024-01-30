/*
Copyright 2024 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package workflow

import (
	"context"
	"encoding/json"

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

// ActivityInput is an option to pass a JSON-serializable input
func ActivityInput(input any) callActivityOption {
	return func(opts *callActivityOptions) error {
		data, err := marshalData(input)
		if err != nil {
			return err
		}
		opts.rawInput = wrapperspb.String(string(data))
		return nil
	}
}

// ActivityRawInput is an option to pass a byte slice as an input
func ActivityRawInput(input string) callActivityOption {
	return func(opts *callActivityOptions) error {
		opts.rawInput = wrapperspb.String(input)
		return nil
	}
}

func marshalData(input any) ([]byte, error) {
	if input == nil {
		return nil, nil
	}
	return json.Marshal(input)
}
