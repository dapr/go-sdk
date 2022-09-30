/*
Copyright 2021 The Dapr Authors
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

package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"

	runtime "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
)

func testBindingHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	if in == nil {
		return nil, errors.New("nil event")
	}
	return in.Data, nil
}

func TestListInputBindings(t *testing.T) {
	server := getTestServer()
	err := server.AddBindingInvocationHandler("test1", testBindingHandler)
	assert.NoError(t, err)
	err = server.AddBindingInvocationHandler("test2", testBindingHandler)
	assert.NoError(t, err)
	resp, err := server.ListInputBindings(context.Background(), &empty.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Lenf(t, resp.Bindings, 2, "expected 2 handlers")
}

func TestBindingForErrors(t *testing.T) {
	server := getTestServer()
	err := server.AddBindingInvocationHandler("", nil)
	assert.Errorf(t, err, "expected error on nil method name")

	err = server.AddBindingInvocationHandler("test", nil)
	assert.Errorf(t, err, "expected error on nil method handler")
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestBinding$
func TestBinding(t *testing.T) {
	ctx := context.Background()
	methodName := "test"

	server := getTestServer()
	err := server.AddBindingInvocationHandler(methodName, testBindingHandler)
	assert.Nil(t, err)
	startTestServer(server)

	t.Run("binding without event", func(t *testing.T) {
		_, err := server.OnBindingEvent(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("binding event for wrong method", func(t *testing.T) {
		in := &runtime.BindingEventRequest{Name: "invalid"}
		_, err := server.OnBindingEvent(ctx, in)
		assert.Error(t, err)
	})

	t.Run("binding event without data", func(t *testing.T) {
		in := &runtime.BindingEventRequest{Name: methodName}
		out, err := server.OnBindingEvent(ctx, in)
		assert.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("binding event with data", func(t *testing.T) {
		data := "hello there"
		in := &runtime.BindingEventRequest{
			Name: methodName,
			Data: []byte(data),
		}
		out, err := server.OnBindingEvent(ctx, in)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, data, string(out.Data))
	})

	t.Run("binding event with metadata", func(t *testing.T) {
		in := &runtime.BindingEventRequest{
			Name:     methodName,
			Metadata: map[string]string{"k1": "v1", "k2": "v2"},
		}
		out, err := server.OnBindingEvent(ctx, in)
		assert.NoError(t, err)
		assert.NotNil(t, out)
	})

	stopTestServer(t, server)
}
