package grpc

import (
	"context"
	"fmt"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

// AddTopicEventHandler appends provided event handler with topic name to the service
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn func(ctx context.Context, e *common.TopicEvent) error) error {
	if sub == nil {
		return errors.New("subscription required")
	}
	if sub.Topic == "" {
		return errors.New("topic name required")
	}

	s.topicSubscriptions[sub.Topic] = &topicEventHandler{
		Subscription: sub,
		fn:           fn,
	}
	return nil
}

// ListTopicSubscriptions is called by Dapr to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *Server) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
	subs := make([]*pb.TopicSubscription, 0)
	for k, v := range s.topicSubscriptions {
		sub := &pb.TopicSubscription{
			Topic:    k,
			Metadata: v.Metadata,
		}
		subs = append(subs, sub)
	}

	return &pb.ListTopicSubscriptionsResponse{
		Subscriptions: subs,
	}, nil
}

// OnTopicEvent fired whenever a message has been published to a topic that has been subscribed. Dapr sends published messages in a CloudEvents 0.3 envelope.
func (s *Server) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*pb.TopicEventResponse, error) {
	if in == nil {
		return nil, errors.New("nil event request")
	}
	if in.Topic == "" {
		return nil, errors.New("topic event request has no topic name")
	}
	if h, ok := s.topicSubscriptions[in.Topic]; ok {
		e := &common.TopicEvent{
			ID:              in.Id,
			Source:          in.Source,
			Type:            in.Type,
			SpecVersion:     in.SpecVersion,
			DataContentType: in.DataContentType,
			Data:            in.Data,
			Topic:           in.Topic,
		}
		err := h.fn(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error handling topic event: %s", in.Topic)
		}
		return &pb.TopicEventResponse{}, nil
	}
	return &pb.TopicEventResponse{}, fmt.Errorf("topic not configured: %s", in.Topic)
}
