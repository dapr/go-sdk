package http

// TopicEvent is the content of the inbound topic message
type TopicEvent struct {
	// ID identifies the event.
	ID string `json:"id"`

	// Source identifies the context in which an event happened.
	Source string `json:"source"`

	// The type of event related to the originating occurrence.
	Type string `json:"type"`

	// The version of the CloudEvents specification.
	SpecVersion string `json:"specversion"`

	// The content type of data value.
	DataContentType string `json:"datacontenttype"`

	// The content of the event.
	Data interface{} `json:"data"`

	// The pubsub topic which publisher sent to.
	Topic string `json:"-"`

	// Cloud event subject
	Subject string `json:"subject"`
}

// InvocationEvent represents the input and output of binding invocation
type InvocationEvent struct {

	// ContentType of the Data
	ContentType string

	// Data is the payload that the input bindings sent.
	Data []byte
}
