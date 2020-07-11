package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	cpb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/server/event"
	"google.golang.org/grpc"
)

// Server is the gRPC Dapr server
type Server interface {
	Start() error
	Stop() error
	AddInvocationHandler(method string, fn func(ctx context.Context, contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte))
	OnInvoke(ctx context.Context, in *cpb.InvokeRequest) (*cpb.InvokeResponse, error)
	AddTopicEventHandler(topic string, fn func(ctx context.Context, event *event.TopicEvent) error)
	ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error)
	OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*empty.Empty, error)
	AddBindingEventHandler(name string, fn func(ctx context.Context, in *event.BindingEvent) error)
	ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error)
	OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error)
}

// NewServer creates new Server
func NewServer(port string) (server Server, err error) {
	address := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = errors.Wrapf(err, "failed to TCP listen on: %s", address)
		return
	}
	server = NewServerWithListener(lis)
	return
}

// NewServerWithListener creates new Server with specific listener
func NewServerWithListener(lis net.Listener) Server {
	return &ServerImp{
		listener:           lis,
		invokeHandlers:     make(map[string]func(ctx context.Context, contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte)),
		topicSubscriptions: make(map[string]func(ctx context.Context, event *event.TopicEvent) error),
		bindingHandlers:    make(map[string]func(ctx context.Context, in *event.BindingEvent) error),
	}
}

// ServerImp is the gRPC server implementation for Dapr
type ServerImp struct {
	listener           net.Listener
	invokeHandlers     map[string]func(ctx context.Context, contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte)
	topicSubscriptions map[string]func(ctx context.Context, event *event.TopicEvent) error
	bindingHandlers    map[string]func(ctx context.Context, in *event.BindingEvent) error
}

// Start registers the server and starts it
func (s *ServerImp) Start() error {
	gs := grpc.NewServer()
	pb.RegisterAppCallbackServer(gs, s)
	return gs.Serve(s.listener)
}

// Stop stops the previously started server
func (s *ServerImp) Stop() error {
	return s.listener.Close()
}
