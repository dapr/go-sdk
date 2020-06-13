package client

import (
	"context"
	"net"
	"testing"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	commonv1pb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

func getTestClient(ctx context.Context) (*Client, func()) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	s := grpc.NewServer()
	pb.RegisterAppCallbackServer(s, &testServer{})
	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())

	closer := func() {
		listener.Close()
		s.Stop()
	}

	client, err := NewClientWithConnection(conn)
	if err != nil {
		panic(err)
	}

	return client, closer
}

func TestNewClientWithConnection(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx)
	assert.NotNil(t, closer)
	defer closer()
	assert.NotNil(t, client)
}

func TestNewClientWithoutArgs(t *testing.T) {
	_, err := NewClientWithPort("")
	assert.NotNil(t, err)
	_, err = NewClientWithAddress("")
	assert.NotNil(t, err)
}

type testServer struct {
}

func (s *testServer) EchoMethod() string {
	return "pong"
}

func (s *testServer) OnInvoke(ctx context.Context, in *commonv1pb.InvokeRequest) (*commonv1pb.InvokeResponse, error) {
	var response string

	switch in.Method {
	case "EchoMethod":
		response = s.EchoMethod()
	}

	return &commonv1pb.InvokeResponse{
		ContentType: "text/plain; charset=UTF-8",
		Data:        &any.Any{Value: []byte(response)},
	}, nil
}

func (s *testServer) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
	return &pb.ListTopicSubscriptionsResponse{
		Subscriptions: []*pb.TopicSubscription{
			{Topic: "TopicA"},
		},
	}, nil
}

func (s *testServer) ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error) {
	return &pb.ListInputBindingsResponse{
		Bindings: []string{"storage"},
	}, nil
}

func (s *testServer) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	return &pb.BindingEventResponse{}, nil
}

func (s *testServer) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
