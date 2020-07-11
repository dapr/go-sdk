package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/server/event"
)

// START TOPIC SUB

// AddTopicEventHandler adds provided topic to the list of server subscriptions
func (s *ServerImp) AddTopicEventHandler(topic string, fn func(ctx context.Context, event *event.TopicEvent) error) {
	s.topicSubscriptions[topic] = fn
}

// ListTopicSubscriptions is called by Dapr to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *ServerImp) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
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
func (s *ServerImp) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*empty.Empty, error) {
	if val, ok := s.topicSubscriptions[in.Topic]; ok {
		e := &event.TopicEvent{
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
