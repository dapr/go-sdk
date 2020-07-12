package http

import "time"

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

	// TIme is the event of time
	Time time.Time `json:"time"`

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

	// ContentType of the Data
	ContentType string

	// Data is the payload that the input bindings sent.
	Data []byte
}

// Subscription represents single topic subscription
type Subscription struct {
	Topic string `json:"topic"`
	Route string `json:"route"`
}
