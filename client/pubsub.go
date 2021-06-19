package client

import (
	"context"
	"encoding/json"
	"log"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEventOption is the type for the functional option.
type PublishEventOption func(*pb.PublishEventRequest)

// PublishEvent publishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, pubsubName, topicName string, data interface{}, opts ...PublishEventOption) error {
	if pubsubName == "" {
		return errors.New("pubsubName name required")
	}
	if topicName == "" {
		return errors.New("topic name required")
	}

	envelope := &pb.PublishEventRequest{
		PubsubName: pubsubName,
		Topic:      topicName,
	}
	for _, opt := range opts {
		opt(envelope)
	}

	if data != nil {
		switch d := data.(type) {
		case []byte:
			envelope.Data = d
		case string:
			envelope.Data = []byte(d)
		default:
			var err error
			envelope.DataContentType = "application/json"
			envelope.Data, err = json.Marshal(d)
			if err != nil {
				return errors.WithMessage(err, "error serializing input struct")
			}
		}
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), envelope)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topicName)
	}

	return nil
}

// PublishEventWithContentType can be passed as option to PublishEvent to set an explicit Content-Type.
func PublishEventWithContentType(contentType string) PublishEventOption {
	return func(e *pb.PublishEventRequest) {
		e.DataContentType = contentType
	}
}

// PublishEventWithMetadata can be passed as option to PublishEvent to set metadata.
func PublishEventWithMetadata(metadata map[string]string) PublishEventOption {
	return func(e *pb.PublishEventRequest) {
		e.Metadata = metadata
	}
}

// PublishEventfromCustomContent serializes an struct and publishes its contents as data (JSON) onto topic in specific pubsub component.
// DEPRECATED: This method is deprecated and will be removed in a future version of the SDK. Please use `PublishEvent` instead.
func (c *GRPCClient) PublishEventfromCustomContent(ctx context.Context, pubsubName, topicName string, data interface{}) error {
	log.Println("DEPRECATED: client.PublishEventfromCustomContent is deprecated and will be removed in a future version of the SDK. Please use `PublishEvent` instead.")

	// Perform the JSON marshaling here just in case someone passed a []byte or string as data
	enc, err := json.Marshal(data)
	if err != nil {
		return errors.WithMessage(err, "error serializing input struct")
	}

	return c.PublishEvent(ctx, pubsubName, topicName, enc, PublishEventWithContentType("application/json"))
}
