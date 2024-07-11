/*
Copyright 2024 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"strings"
	"sync"
	"sync/atomic"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
)

type SubscriptionHandleFunction func(event *common.TopicEvent) common.SubscriptionResponseStatus

type SubscriptionOptions struct {
	PubsubName      string
	Topic           string
	DeadLetterTopic *string
	Metadata        map[string]string
}

type Subscription struct {
	stream pb.Dapr_SubscribeTopicEventsAlpha1Client
	// lock locks concurrent writes to subscription stream.
	lock   sync.Mutex
	closed atomic.Bool
}

type SubscriptionMessage struct {
	*common.TopicEvent
	sub *Subscription
}

func (c *GRPCClient) Subscribe(ctx context.Context, opts SubscriptionOptions) (*Subscription, error) {
	stream, err := c.subscribeInitialRequest(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Subscription{
		stream: stream,
	}, nil
}

func (c *GRPCClient) SubscribeWithHandler(ctx context.Context, opts SubscriptionOptions, handler SubscriptionHandleFunction) (func() error, error) {
	s, err := c.Subscribe(ctx, opts)
	if err != nil {
		return nil, err
	}

	go func() {
		defer s.Close()

		for {
			msg, err := s.Receive()
			if err != nil {
				if !s.closed.Load() {
					logger.Printf("Error receiving messages from subscription pubsub=%s topic=%s, closing subscription: %s",
						opts.PubsubName, opts.Topic, err)
				}
				return
			}

			go func() {
				if err := msg.respondStatus(handler(msg.TopicEvent)); err != nil {
					logger.Printf("Error responding to topic with event status pubsub=%s topic=%s message_id=%s: %s",
						opts.PubsubName, opts.Topic, msg.ID, err)
				}
			}()
		}
	}()

	return s.Close, nil
}

func (s *Subscription) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return errors.New("subscription already closed")
	}

	return s.stream.CloseSend()
}

func (s *Subscription) Receive() (*SubscriptionMessage, error) {
	event, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}

	data := any(event.GetData())
	if len(event.GetData()) > 0 {
		mediaType, _, err := mime.ParseMediaType(event.GetDataContentType())
		if err == nil {
			var v interface{}
			switch mediaType {
			case "application/json":
				if err := json.Unmarshal(event.GetData(), &v); err == nil {
					data = v
				}
			case "text/plain":
				// Assume UTF-8 encoded string.
				data = string(event.GetData())
			default:
				if strings.HasPrefix(mediaType, "application/") &&
					strings.HasSuffix(mediaType, "+json") {
					if err := json.Unmarshal(event.GetData(), &v); err == nil {
						data = v
					}
				}
			}
		}
	}

	topicEvent := &common.TopicEvent{
		ID:              event.GetId(),
		Source:          event.GetSource(),
		Type:            event.GetType(),
		SpecVersion:     event.GetSpecVersion(),
		DataContentType: event.GetDataContentType(),
		Data:            data,
		RawData:         event.GetData(),
		Topic:           event.GetTopic(),
		PubsubName:      event.GetPubsubName(),
	}

	return &SubscriptionMessage{
		sub:        s,
		TopicEvent: topicEvent,
	}, nil
}

func (s *SubscriptionMessage) Success() error {
	return s.respond(pb.TopicEventResponse_SUCCESS)
}

func (s *SubscriptionMessage) Retry() error {
	return s.respond(pb.TopicEventResponse_RETRY)
}

func (s *SubscriptionMessage) Drop() error {
	return s.respond(pb.TopicEventResponse_DROP)
}

func (s *SubscriptionMessage) respondStatus(status common.SubscriptionResponseStatus) error {
	var statuspb pb.TopicEventResponse_TopicEventResponseStatus
	switch status {
	case common.SubscriptionResponseStatusSuccess:
		statuspb = pb.TopicEventResponse_SUCCESS
	case common.SubscriptionResponseStatusRetry:
		statuspb = pb.TopicEventResponse_RETRY
	case common.SubscriptionResponseStatusDrop:
		statuspb = pb.TopicEventResponse_DROP
	default:
		return fmt.Errorf("unknown status, expected one of %s, %s, %s: %s",
			common.SubscriptionResponseStatusSuccess, common.SubscriptionResponseStatusRetry,
			common.SubscriptionResponseStatusDrop, status)
	}

	return s.respond(statuspb)
}

func (s *SubscriptionMessage) respond(status pb.TopicEventResponse_TopicEventResponseStatus) error {
	s.sub.lock.Lock()
	defer s.sub.lock.Unlock()

	return s.sub.stream.Send(&pb.SubscribeTopicEventsRequestAlpha1{
		SubscribeTopicEventsRequestType: &pb.SubscribeTopicEventsRequestAlpha1_EventResponse{
			EventResponse: &pb.SubscribeTopicEventsResponseAlpha1{
				Id:     s.ID,
				Status: &pb.TopicEventResponse{Status: status},
			},
		},
	})
}

func (c *GRPCClient) subscribeInitialRequest(ctx context.Context, opts SubscriptionOptions) (pb.Dapr_SubscribeTopicEventsAlpha1Client, error) {
	if len(opts.PubsubName) == 0 {
		return nil, errors.New("pubsub name required")
	}

	if len(opts.Topic) == 0 {
		return nil, errors.New("topic required")
	}

	stream, err := c.protoClient.SubscribeTopicEventsAlpha1(ctx)
	if err != nil {
		return nil, err
	}

	err = stream.Send(&pb.SubscribeTopicEventsRequestAlpha1{
		SubscribeTopicEventsRequestType: &pb.SubscribeTopicEventsRequestAlpha1_InitialRequest{
			InitialRequest: &pb.SubscribeTopicEventsInitialRequestAlpha1{
				PubsubName: opts.PubsubName, Topic: opts.Topic,
				Metadata: opts.Metadata, DeadLetterTopic: opts.DeadLetterTopic,
			},
		},
	})
	if err != nil {
		return nil, errors.Join(err, stream.CloseSend())
	}

	return stream, nil
}
