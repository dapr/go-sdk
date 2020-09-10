package common

import "context"

// Service represents Dapr callback service
type Service interface {
	// AddServiceInvocationHandler appends provided service invocation handler with its name to the service.
	AddServiceInvocationHandler(name string, fn func(ctx context.Context, in *InvocationEvent) (out *Content, err error)) error
	// AddTopicEventHandler appends provided event handler with its topic and optional metadata to the service.
	// Note, retries are only considered when there is an error. Lack of error is considered as a success
	AddTopicEventHandler(sub *Subscription, fn func(ctx context.Context, e *TopicEvent) (retry bool, err error)) error
	// AddBindingInvocationHandler appends provided binding invocation handler with its name to the service.
	AddBindingInvocationHandler(name string, fn func(ctx context.Context, in *BindingEvent) (out []byte, err error)) error
	// Start starts service.
	Start() error
	// Stop stops the previously started service.
	Stop() error
}
