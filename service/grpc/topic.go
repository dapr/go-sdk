package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

// AddTopicEventHandler appends provided event handler with topic name to the service
func (s *ServiceImp) AddTopicEventHandler(topic string, fn func(ctx context.Context, e *TopicEvent) error) error {
	if topic == "" {
		return fmt.Errorf("topic name required")
	}
	s.topicSubscriptions[topic] = fn
	return nil
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
	if in == nil {
		return nil, errors.New("nil event request")
	}
	if in.Topic == "" {
		return nil, errors.New("topic event request has no topic name")
	}
	if fn, ok := s.topicSubscriptions[in.Topic]; ok {
		e := &TopicEvent{
			Topic:           in.Topic,
			Data:            in.Data,
			DataContentType: in.DataContentType,
			ID:              in.Id,
			Source:          in.Source,
			SpecVersion:     in.SpecVersion,
			Type:            in.Type,
		}
		err := fn(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error handling topic event: %s", in.Topic)
		}
		return &empty.Empty{}, nil
	}
	return &empty.Empty{}, fmt.Errorf("topic not configured: %s", in.Topic)
}
