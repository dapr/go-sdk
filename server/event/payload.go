package event

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
	Data []byte `json:"data"`

	// The pubsub topic which publisher sent to.
	Topic string

	// Cloud event subject
	Subject string `json:"subject"`
}

// BindingEvent represents the input and output of binding invocation
type BindingEvent struct {

	// Name of the input binding component.
	Name string

	// Data is the payload that the input bindings sent.
	Data []byte

	// Metadata is set by the input binging components.
	Metadata map[string]string
}
