package http

import (
	"context"
	"net/http"
)

// Service represents Dapr callback service
type Service interface {
	// AddServiceInvocationHandler appends provided service invocation handler with its name to the service.
	AddServiceInvocationHandler(name string, fn func(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error)) error
	// AddTopicEventHandler appends provided event handler with it's topic to the service
	AddTopicEventHandler(topic, route string, fn func(ctx context.Context, e *TopicEvent) error) error
	// AddBindingInvocationHandler appends provided binding invocation handler with its route to the service
	AddBindingInvocationHandler(route string, fn func(ctx context.Context, in *BindingEvent) (out []byte, err error)) error
	// Start starts service
	Start() error
}

// TopicEvent is the content of the inbound topic message
type TopicEvent struct {
	// ID identifies the event.
	ID string `json:"id"`
	// The version of the CloudEvents specification.
	SpecVersion string `json:"specversion"`
	// The type of event related to the originating occurrence.
	Type string `json:"type"`
	// Source identifies the context in which an event happened.
	Source string `json:"source"`
	// The content type of data value.
	DataContentType string `json:"datacontenttype"`
	// The content of the event.
	Data []byte `json:"data"`
	// Cloud event subject
	Subject string `json:"subject"`
	// The pubsub topic which publisher sent to.
	Topic string `json:"topic"`
}

// InvocationEvent represents the input and output of binding invocation
type InvocationEvent struct {
	// Data is the payload that the input bindings sent.
	Data []byte `json:"data"`
	// ContentType of the Data
	ContentType string `json:"contentType"`
	// DataTypeURL is the resource URL that uniquely identifies the type of the serialized
	DataTypeURL string `json:"typeUrl,omitempty"`
}

// BindingEvent represents the binding event handler input
type BindingEvent struct {
	// Data is the input bindings sent
	Data []byte `json:"data"`
	// Metadata is the input binging components
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Subscription represents single topic subscription
type Subscription struct {
	// Topic is the name of the topic
	Topic string `json:"topic"`
	// Route is the route of the handler where topic events should be published
	Route string `json:"route"`
}

// NewService creates new Service
func NewService(address string) Service {
	return newService(address)
}

func newService(address string) *ServiceImp {
	return &ServiceImp{
		address:            address,
		mux:                http.NewServeMux(),
		topicSubscriptions: make([]*Subscription, 0),
	}
}

// ServiceImp is the HTTP server wrapping mux many Dapr helpers
type ServiceImp struct {
	address            string
	mux                *http.ServeMux
	topicSubscriptions []*Subscription
}

// Start starts the HTTP handler. Blocks while serving
func (s *ServiceImp) Start() error {
	s.registerSubscribeHandler()
	server := http.Server{
		Addr:    s.address,
		Handler: s.mux,
	}
	return server.ListenAndServe()
}

func optionsHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
			w.Header().Set("Allow", "POST,OPTIONS")
		} else {
			h.ServeHTTP(w, r)
		}
	}
}
