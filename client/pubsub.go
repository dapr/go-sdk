package client

import (
	"context"
	"encoding/json"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

type PublishEventRequest struct {
	// The name of the pubsub component
	PubsubName string
	// The pubsub topic
	Topic string
	// The data which will be published to topic.
	Data []byte
	// The content type for the data (optional).
	DataContentType string
	// The metadata passing to pub components
	//
	// metadata property:
	// - key : the key of the message.
	Metadata map[string]string
}

// Publish publishes custom event.
func (c *GRPCClient) Publish(ctx context.Context, in *PublishEventRequest) error {
	if in.PubsubName == "" {
		return errors.New("PublishEventRequest.PubsubName required")
	}
	if in.Topic == "" {
		return errors.New("PublishEventRequest.Topic required")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName:      in.PubsubName,
		Topic:           in.Topic,
		Data:            in.Data,
		DataContentType: in.DataContentType,
		Metadata:        in.Metadata,
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", in.Topic)
	}

	return nil
}

// PublishEvent publishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, pubsubName, topicName string, data []byte) error {
	if pubsubName == "" {
		return errors.New("pubsubName name required")
	}
	if topicName == "" {
		return errors.New("topic name required")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName: pubsubName,
		Topic:      topicName,
		Data:       data,
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topicName)
	}

	return nil
}

// PublishEventfromCustomContent serializes an struct and publishes its contents as data (JSON) onto topic in specific pubsub component.
func (c *GRPCClient) PublishEventfromCustomContent(ctx context.Context, pubsubName, topicName string, data interface{}) error {
	if pubsubName == "" {
		return errors.New("pubsubName name required")
	}
	if topicName == "" {
		return errors.New("topic name required")
	}

	bytes, err := json.Marshal(data)

	if err != nil {
		return errors.WithMessage(err, "error serializing input struct")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName: pubsubName,
		Topic:      topicName,
		Data:       bytes,
	}

	_, err = c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)

	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topicName)
	}

	return nil
}
