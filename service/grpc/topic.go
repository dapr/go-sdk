package grpc

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

// AddTopicEventHandler appends provided event handler with topic name to the service
func (s *Server) AddTopicEventHandler(component, topic string, fn func(ctx context.Context, e *TopicEvent) error) error {
	if topic == "" {
		return fmt.Errorf("topic name required")
	}
	s.topicSubscriptions[topic] = &topicEventHandler{
		component: component,
		topic:     topic,
		fn:        fn,
		meta:      map[string]string{},
	}
	return nil
}

// AddTopicEventHandlerWithMetadata appends provided event handler with topic name and metadata to the service
func (s *Server) AddTopicEventHandlerWithMetadata(component, topic string, m map[string]string, fn func(ctx context.Context, e *TopicEvent) error) error {
	if topic == "" {
		return fmt.Errorf("topic name required")
	}
	s.topicSubscriptions[topic] = &topicEventHandler{
		component: component,
		topic:     topic,
		fn:        fn,
		meta:      m,
	}
	return nil
}

// ListTopicSubscriptions is called by Dapr to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *Server) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
	subs := make([]*pb.TopicSubscription, 0)
	for _, v := range s.topicSubscriptions {
		sub := &pb.TopicSubscription{
			PubsubName: v.component,
			Topic:      v.topic,
			Metadata:   v.meta,
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
		e := &TopicEvent{
			ID:              in.Id,
			Source:          in.Source,
			Type:            in.Type,
			SpecVersion:     in.SpecVersion,
			DataContentType: in.DataContentType,
			Data:            in.Data,
			Topic:           in.Topic,
			PubsubName:      in.PubsubName,
		}
		err := h.fn(ctx, e)
		if err != nil {
			return nil, errors.Wrapf(err, "error handling topic event: %s", in.Topic)
		}
		return &pb.TopicEventResponse{}, nil
	}
	return &pb.TopicEventResponse{}, fmt.Errorf("topic not configured: %s", in.Topic)
}
