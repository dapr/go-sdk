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
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)) error {
	if sub == nil {
		return errors.New("subscription required")
	}
	if sub.Topic == "" {
		return errors.New("topic name required")
	}
	if sub.PubsubName == "" {
		return errors.New("pub/sub name required")
	}
	if fn == nil {
		return fmt.Errorf("topic handler required")
	}
	key := fmt.Sprintf("%s-%s", sub.PubsubName, sub.Topic)
	s.topicSubscriptions[key] = &topicEventHandler{
		component: sub.PubsubName,
		topic:     sub.Topic,
		fn:        fn,
		meta:      sub.Metadata,
	}
	return nil
}

// ListTopicSubscriptions is called by Dapr to get the list of topics in a pubsub component the app wants to subscribe to.
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

// OnTopicEvent fired whenever a message has been published to a topic that has been subscribed.
// Dapr sends published messages in a CloudEvents v1.0 envelope.
func (s *Server) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*pb.TopicEventResponse, error) {
	if in == nil || in.Topic == "" || in.PubsubName == "" {
		// this is really Dapr issue more than the event request format.
		// since Dapr will not get updated until long after this event expires, just drop it
		return &pb.TopicEventResponse{Status: pb.TopicEventResponse_DROP}, errors.New("pub/sub and topic names required")
	}
	key := fmt.Sprintf("%s-%s", in.PubsubName, in.Topic)
	if h, ok := s.topicSubscriptions[key]; ok {
		e := &common.TopicEvent{
			ID:              in.Id,
			Source:          in.Source,
			Type:            in.Type,
			SpecVersion:     in.SpecVersion,
			DataContentType: in.DataContentType,
			Data:            in.Data,
			Topic:           in.Topic,
			PubsubName:      in.PubsubName,
		}
		retry, err := h.fn(ctx, e)
		if err == nil {
			return &pb.TopicEventResponse{Status: pb.TopicEventResponse_SUCCESS}, nil
		}
		if retry {
			return &pb.TopicEventResponse{Status: pb.TopicEventResponse_RETRY}, err
		}
		return &pb.TopicEventResponse{Status: pb.TopicEventResponse_DROP}, err
	}
	return &pb.TopicEventResponse{Status: pb.TopicEventResponse_RETRY}, fmt.Errorf(
		"pub/sub and topic combination not configured: %s/%s",
		in.PubsubName, in.Topic,
	)
}
