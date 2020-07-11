package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/server/event"
	"github.com/stretchr/testify/assert"
)

// go test -v -count=1 -run TestBinding ./server/grpc
func TestBinding(t *testing.T) {
	methodName := "test"
	data := "hello there"

	server := getTestServer()
	server.AddBindingEventHandler(methodName, bindingHandler)
	startTestServer(server)

	ctx := context.Background()
	in := &runtime.BindingEventRequest{
		Name: methodName,
		Data: []byte(data),
		Metadata: map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		},
	}

	out, err := server.OnBindingEvent(ctx, in)
	assert.NoError(t, err)
	assert.NotNil(t, out)

	stopTestServer(t, server)
}

func bindingHandler(ctx context.Context, in *event.BindingEvent) error {
	if in == nil {
		return errors.New("nil event")
	}
	return nil
}
