package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEvent publishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, pubsubName, topicName string, data []byte) error {
	return c.PublishEventWithMetadata(ctx, pubsubName, topicName, data, nil)
}

// PublishEventWithMetadata publishes data onto specific pubsub topic with support for metadata.
func (c *GRPCClient) PublishEventWithMetadata(ctx context.Context, pubsubName, topicName string, data []byte, metadata map[string]string) error {
	if pubsubName == "" {
		return errors.New("pubsubName name required")
	}
	if topicName == "" {
		return errors.New("topic name required")
	}

	envelope := &pb.PublishEventRequest{
		PubsubName: pubsubName,
		Topic:      topicName,
		Data:       data,
		Metadata:   metadata,
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), envelope)
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
		PubsubName:      pubsubName,
		Topic:           topicName,
		Data:            bytes,
		DataContentType: "application/json",
	}

	_, err = c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)

	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topicName)
	}

	return nil
}
