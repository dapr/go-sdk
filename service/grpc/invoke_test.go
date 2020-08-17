package grpc

import (
	"context"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/common/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

func testInvokeHandler(ctx context.Context, in *InvocationEvent) (out *Content, err error) {
	if in == nil {
		return
	}
	out = &Content{
		ContentType: in.ContentType,
		Data:        in.Data,
	}
	return
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestInvoke$
func TestInvoke(t *testing.T) {
	methodName := "test"
	ctx := context.Background()

	server := getTestServer()
	err := server.AddServiceInvocationHandler(methodName, testInvokeHandler)
	assert.Nil(t, err)
	startTestServer(server)

	t.Run("invoke without request", func(t *testing.T) {
		_, err := server.OnInvoke(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("invoke request with invalid method name", func(t *testing.T) {
		in := &common.InvokeRequest{Method: "invalid"}
		_, err := server.OnInvoke(ctx, in)
		assert.Error(t, err)
	})

	t.Run("invoke request without data", func(t *testing.T) {
		in := &common.InvokeRequest{Method: methodName}
		_, err := server.OnInvoke(ctx, in)
		assert.NoError(t, err)
	})

	t.Run("invoke request with data", func(t *testing.T) {
		data := "hello there"
		dataContentType := "text/plain"
		in := &common.InvokeRequest{Method: methodName}
		in.Data = &anypb.Any{Value: []byte(data)}
		in.ContentType = dataContentType
		out, err := server.OnInvoke(ctx, in)
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, dataContentType, out.ContentType)
		assert.Equal(t, data, string(out.Data.Value))
	})

	stopTestServer(t, server)
}
