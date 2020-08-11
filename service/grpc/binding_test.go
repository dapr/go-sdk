package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/stretchr/testify/assert"
)

func testBindingHandler(ctx context.Context, in *BindingEvent) (out []byte, err error) {
	if in == nil {
		return nil, errors.New("nil event")
	}
	return in.Data, nil
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestBinding$
func TestBinding(t *testing.T) {
	t.Parallel()

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
