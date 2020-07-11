package grpc

import (
	"context"
	"net"

	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"google.golang.org/grpc"
)

// Service is the gRPC Dapr service
type Service interface {
	Start() error
	AddInvocationHandler(method string, fn func(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error))
	AddTopicEventHandler(topic string, fn func(ctx context.Context, event *TopicEvent) error)
	AddBindingEventHandler(name string, fn func(ctx context.Context, in *BindingEvent) (out []byte, err error))
	Stop() error
}

// NewService creates new Service
func NewService(address string) (server Service, err error) {
	if address == "" {
		return nil, errors.New("nil address")
	}
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = errors.Wrapf(err, "failed to TCP listen on: %s", address)
		return
	}
	server = newService(lis)
	return
}

// NewServiceWithListener creates new Service with specific listener
func NewServiceWithListener(lis net.Listener) Service {
	return newService(lis)
}

func newService(lis net.Listener) *ServiceImp {
	return &ServiceImp{
		listener:           lis,
		invokeHandlers:     make(map[string]func(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error)),
		topicSubscriptions: make(map[string]func(ctx context.Context, e *TopicEvent) error),
		bindingHandlers:    make(map[string]func(ctx context.Context, in *BindingEvent) (out []byte, err error)),
	}
}

// ServiceImp is the gRPC service implementation for Dapr
type ServiceImp struct {
	listener           net.Listener
	invokeHandlers     map[string]func(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error)
	topicSubscriptions map[string]func(ctx context.Context, e *TopicEvent) error
	bindingHandlers    map[string]func(ctx context.Context, in *BindingEvent) (out []byte, err error)
}

// Start registers the server and starts it
func (s *ServiceImp) Start() error {
	gs := grpc.NewServer()
	pb.RegisterAppCallbackServer(gs, s)
	return gs.Serve(s.listener)
}

// Stop stops the previously started server
func (s *ServiceImp) Stop() error {
	return s.listener.Close()
}
