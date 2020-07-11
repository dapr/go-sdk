package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

// TopicEvent is the content of the inbound topic message
type TopicEvent struct {
	// ID identifies the event.
	ID string
	// Source identifies the context in which an event happened.
	Source string
	// The type of event related to the originating occurrence.
	Type string
	// The version of the CloudEvents specification.
	SpecVersion string
	// The content type of data value.
	DataContentType string
	// The content of the event.
	Data []byte
	// The pubsub topic which publisher sent to.
	Topic string
	// Cloud event subject
	Subject string
}

// AddTopicEventHandler adds provided topic to the list of server subscriptions
func (s *ServiceImp) AddTopicEventHandler(topic string, fn func(ctx context.Context, e *TopicEvent) error) {
	s.topicSubscriptions[topic] = fn
}

// ListTopicSubscriptions is called by Dapr to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *ServiceImp) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
	subs := make([]*pb.TopicSubscription, 0)
	for k := range s.topicSubscriptions {
		sub := &pb.TopicSubscription{
			Topic: k,
		}
		subs = append(subs, sub)
	}

	return &pb.ListTopicSubscriptionsResponse{
		Subscriptions: subs,
	}, nil
}

// OnTopicEvent fired whenever a message has been published to a topic that has been subscribed. Dapr sends published messages in a CloudEvents 0.3 envelope.
func (s *ServiceImp) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*empty.Empty, error) {
	if val, ok := s.topicSubscriptions[in.Topic]; ok {
		e := &TopicEvent{
			Topic:           in.Topic,
			Data:            in.Data,
			DataContentType: in.DataContentType,
			ID:              in.Id,
			Source:          in.Source,
			SpecVersion:     in.SpecVersion,
			Type:            in.Type,
		}
		err := val(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error handling topic event: %s", in.Topic)
		}
	}
	return &empty.Empty{}, nil
}
