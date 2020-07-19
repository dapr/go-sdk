package grpc

import (
	"context"
	"net"

	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"google.golang.org/grpc"
)

// Service represents Dapr callback service
type Service interface {
	// AddServiceInvocationHandler appends provided service invocation handler with its name to the service.
	AddServiceInvocationHandler(name string, fn func(ctx context.Context, in *InvocationEvent) (out *Content, err error)) error
	// AddTopicEventHandler appends provided event handler with it's topic to the service
	AddTopicEventHandler(topic string, fn func(ctx context.Context, e *TopicEvent) error) error
	// AddBindingInvocationHandler appends provided binding invocation handler with its name to the service
	AddBindingInvocationHandler(name string, fn func(ctx context.Context, in *BindingEvent) (out []byte, err error)) error
	// Start starts service
	Start() error
	// Stop stops the previously started service
	Stop() error
}

// TopicEvent is the content of the inbound topic message
type TopicEvent struct {
	// ID identifies the event.
	ID string
	// The version of the CloudEvents specification.
	SpecVersion string
	// The type of event related to the originating occurrence.
	Type string
	// Source identifies the context in which an event happened.
	Source string
	// The content type of data value.
	DataContentType string
	// The content of the event.
	Data interface{}
	// Cloud event subject
	Subject string
	// The pubsub topic which publisher sent to.
	Topic string
}

// InvocationEvent represents the input and output of binding invocation
type InvocationEvent struct {
	// Data is the payload that the input bindings sent.
	Data []byte
	// ContentType of the Data
	ContentType string
	// DataTypeURL is the resource URL that uniquely identifies the type of the serialized.
	DataTypeURL string
	// Verb is the HTTP verb that was used to invoke this service.
	Verb string
	// QueryString is the HTTP query string that was used to invoke this service.
	QueryString map[string]string
}

// Content is a generic data content
type Content struct {
	// Data is the payload that the input bindings sent.
	Data []byte
	// ContentType of the Data
	ContentType string
	// DataTypeURL is the resource URL that uniquely identifies the type of the serialized.
	DataTypeURL string
}

// BindingEvent represents the binding event handler input
type BindingEvent struct {
	// Data is the input bindings sent
	Data []byte
	// Metadata is the input binging components
	Metadata map[string]string
}

// Subscription represents single topic subscription
type Subscription struct {
	// Topic is the name of the topic
	Topic string
	// Route is the route of the handler where topic events should be published
	Route string
}

// NewService creates new Service
func NewService(address string) (s Service, err error) {
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

// NewServiceWithListener creates new Service with specific listener
func NewServiceWithListener(lis net.Listener) Service {
	return newService(lis)
}

func newService(lis net.Listener) *ServiceImp {
	return &ServiceImp{
		listener:           lis,
		invokeHandlers:     make(map[string]func(ctx context.Context, in *InvocationEvent) (out *Content, err error)),
		topicSubscriptions: make(map[string]func(ctx context.Context, e *TopicEvent) error),
		bindingHandlers:    make(map[string]func(ctx context.Context, in *BindingEvent) (out []byte, err error)),
	}
}

// ServiceImp is the gRPC service implementation for Dapr
type ServiceImp struct {
	listener           net.Listener
	invokeHandlers     map[string]func(ctx context.Context, in *InvocationEvent) (out *Content, err error)
	topicSubscriptions map[string]func(ctx context.Context, e *TopicEvent) error
	bindingHandlers    map[string]func(ctx context.Context, in *BindingEvent) (out []byte, err error)
}

// Start registers the server and starts it
func (s *ServiceImp) Start() error {
	gs := grpc.NewServer()
	pb.RegisterAppCallbackServer(gs, s)
	return gs.Serve(s.listener)
}

// Stop stops the previously started service
func (s *ServiceImp) Stop() error {
	return s.listener.Close()
}
