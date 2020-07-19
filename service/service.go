package service

import (
	"context"
)

// Service represents Dapr callback service
type Service interface {
	// AddServiceInvocationHandler appends provided service invocation handler with its name to the service.
	AddServiceInvocationHandler(name string, fn func(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error)) error
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
	Data interface{} `json:"data"`
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
