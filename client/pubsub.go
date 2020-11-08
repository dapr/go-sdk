package client

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// PublishEvent pubishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, component, topic string, in []byte) error {
	if topic == "" {
		return errors.New("topic name required")
	}
	if component == "" {
		return errors.New("component name required")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName: component,
		Topic:      topic,
		Data:       in,
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)
	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topic)
	}

	return nil
}

// PublishEventfromStruct serializes an struct and pubishes its contents as data (JSON) onto topic in specific pubsub component.
func (c *GRPCClient) PublishEventfromStruct(ctx context.Context, component, topic string, in interface{}) error {

	if topic == "" {
		return errors.New("topic name required")
	}
	if component == "" {
		return errors.New("component name required")
	}

	bytes, err := json.Marshal(in)

	if err != nil {
		return errors.WithMessage(err, "error serializing input struct")
	}

	envelop := &pb.PublishEventRequest{
		PubsubName: component,
		Topic:      topic,
		Data:       bytes,
	}

	_, err = c.protoClient.PublishEvent(c.withAuthToken(ctx), envelop)

	if err != nil {
		return errors.Wrapf(err, "error publishing event unto %s topic", topic)
	}

	return nil
}
