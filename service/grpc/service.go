package grpc

import (
	"context"
	"net"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	"github.com/dapr/go-sdk/service/common"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
)

// NewService creates new Service.
func NewService(address string) (s common.Service, err error) {
	if address == "" {
		return nil, errors.New("nil address")
	}
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = errors.Wrapf(err, "failed to TCP listen on: %s", address)
		return
	}
	s = newService(lis)
	return
}

// NewServiceWithListener creates new Service with specific listener.
func NewServiceWithListener(lis net.Listener) common.Service {
	return newService(lis)
}

func newService(lis net.Listener) *Server {
	return &Server{
		listener:           lis,
		invokeHandlers:     make(map[string]func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)),
		topicSubscriptions: make(map[string]*topicEventHandler),
		bindingHandlers:    make(map[string]func(ctx context.Context, in *common.BindingEvent) (out []byte, err error)),
	}
}

// Server is the gRPC service implementation for Dapr.
type Server struct {
	pb.UnimplementedAppCallbackServer
	listener           net.Listener
	invokeHandlers     map[string]func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	topicSubscriptions map[string]*topicEventHandler
	bindingHandlers    map[string]func(ctx context.Context, in *common.BindingEvent) (out []byte, err error)
}

func (s *Server) RegisterActorImplFactory(f actor.Factory, opts ...config.Option) {
	panic("Actor is not supported by gRPC API")
}

type topicEventHandler struct {
	component string
	topic     string
	fn        func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)
	meta      map[string]string
}

// Start registers the server and starts it.
func (s *Server) Start() error {
	gs := grpc.NewServer()
	pb.RegisterAppCallbackServer(gs, s)
	return gs.Serve(s.listener)
}

// Stop stops the previously started service.
func (s *Server) Stop() error {
	return s.listener.Close()
}
