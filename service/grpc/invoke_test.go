package grpc

import (
	"context"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/common/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

// go test -v -count=1 -run TestServer ./server/grpc
func TestInvoke(t *testing.T) {
	t.Parallel()

	methodName := "test"
	data := "hello there"
	dataContentType := "text/plain"

	server := getTestServer()
	server.AddInvocationHandler(methodName, invocationHandler)
	startTestServer(server)

	ctx := context.Background()
	in := &common.InvokeRequest{
		Method:      methodName,
		ContentType: dataContentType,
		Data: &anypb.Any{
			Value: []byte(data),
		},
	}

	out, err := server.OnInvoke(ctx, in)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, dataContentType, out.ContentType)
	assert.Equal(t, data, string(out.Data.Value))

	stopTestServer(t, server)
}

func invocationHandler(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error) {
	out = &InvocationEvent{
		ContentType: in.ContentType,
		Data:        in.Data,
	}
	return
}
