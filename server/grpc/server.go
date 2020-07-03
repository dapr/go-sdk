package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	commonv1pb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/server/event"
	"google.golang.org/grpc"
)

// NewServer creates new Server
func NewServer(port string) (server *Server, err error) {
	address := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = errors.Wrapf(err, "failed to TCP listen on: %s", address)
		return
	}

	server = &Server{
		listener:           lis,
		invokeHandlers:     make(map[string]func(contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte)),
		topicSubscriptions: make(map[string]func(event *event.TopicEvent) error),
		bindingHandlers:    make(map[string]func(in *event.BindingEvent) error),
	}
	return
}

// Server is the gRPC server implementation for Dapr
type Server struct {
	listener           net.Listener
	invokeHandlers     map[string]func(contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte)
	topicSubscriptions map[string]func(event *event.TopicEvent) error
	bindingHandlers    map[string]func(in *event.BindingEvent) error
}

// Start registers the server and starts it
func (s *Server) Start() error {
	gs := grpc.NewServer()
	pb.RegisterAppCallbackServer(gs, s)
	return gs.Serve(s.listener)
}

// START INVOKE

// Invocation

// AddInvocationHandler adds provided handler to the local collection before server start
func (s *Server) AddInvocationHandler(method string, fn func(contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte)) {
	s.invokeHandlers[method] = fn
}

// OnInvoke gets invoked when a remote service has called the app through Dapr
func (s *Server) OnInvoke(ctx context.Context, in *commonv1pb.InvokeRequest) (*commonv1pb.InvokeResponse, error) {
	if val, ok := s.invokeHandlers[in.Method]; ok {
		ct, d := val(in.ContentType, in.Data.Value)
		return &commonv1pb.InvokeResponse{
			ContentType: ct,
			Data:        &any.Any{Value: d},
		}, nil
	}
	return nil, fmt.Errorf("method not implemented: %s", in.Method)
}

// START TOPIC SUB

// AddTopicEventHandler adds provided topic to the list of server subscriptions
func (s *Server) AddTopicEventHandler(topic string, fn func(event *event.TopicEvent) error) {
	s.topicSubscriptions[topic] = fn
}

// ListTopicSubscriptions is called by Dapr to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *Server) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
	subs := make([]*pb.TopicSubscription, 0)
	for k := range s.topicSubscriptions {
		sub := &pb.TopicSubscription{
			Topic: k,
		}
		subs = append(subs, sub)
	}

	return &pb.ListTopicSubscriptionsResponse{
		Subscriptions: subs,
	}, nil
}

// OnTopicEvent fired whenever a message has been published to a topic that has been subscribed. Dapr sends published messages in a CloudEvents 0.3 envelope.
func (s *Server) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*empty.Empty, error) {
	if val, ok := s.topicSubscriptions[in.Topic]; ok {
		e := &event.TopicEvent{
			Topic:           in.Topic,
			Data:            in.Data,
			DataContentType: in.DataContentType,
			ID:              in.Id,
			Source:          in.Source,
			SpecVersion:     in.SpecVersion,
			Type:            in.Type,
		}
		err := val(e)
		if err != nil {
			return nil, errors.Wrapf(err, "error handling topic event: %s", in.Topic)
		}
	}
	return &empty.Empty{}, nil
}

// START BINDING

// AddBindingEventHandler add the provided handler to the server binding halder collection
func (s *Server) AddBindingEventHandler(name string, fn func(in *event.BindingEvent) error) {
	s.bindingHandlers[name] = fn
}

// ListInputBindings is called by Dapr to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage
func (s *Server) ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error) {
	list := make([]string, 0)
	for k := range s.bindingHandlers {
		list = append(list, k)
	}

	return &pb.ListInputBindingsResponse{
		Bindings: list,
	}, nil
}

// OnBindingEvent gets invoked every time a new event is fired from a registered binding. The message carries the binding name, a payload and optional metadata
func (s *Server) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	if val, ok := s.bindingHandlers[in.Name]; ok {
		e := &event.BindingEvent{
			Name:     in.Name,
			Data:     in.Data,
			Metadata: in.Metadata,
		}
		err := val(e)
		if err != nil {
			return nil, errors.Wrapf(err, "error executing %s binding", in.Name)
		}
		return &pb.BindingEventResponse{}, nil
	}

	return nil, fmt.Errorf("binding not implemented: %s", in.Name)
}
