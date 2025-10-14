/*
Copyright 2021 The Dapr Authors
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

package connectrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"strings"

	runtimev1 "buf.build/gen/go/johansja/dapr/protocolbuffers/go/dapr/proto/runtime/v1"
	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// AddTopicEventHandler appends provided event handler with topic name to the service.
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn common.TopicEventHandler) error {
	if fn == nil {
		return errors.New("topic handler required")
	}

	return s.AddTopicEventSubscriber(sub, fn)
}

// AddTopicEventSubscriber appends the provided subscriber to the service.
func (s *Server) AddTopicEventSubscriber(sub *common.Subscription, subscriber common.TopicEventSubscriber) error {
	if sub == nil {
		return errors.New("subscription required")
	}

	return s.topicRegistrar.AddSubscription(sub, subscriber)
}

// ListTopicSubscriptions is called by Dapr to get the list of topics in a pubsub component the app wants to subscribe to.
func (s *Server) ListTopicSubscriptions(ctx context.Context, in *connect.Request[emptypb.Empty]) (*connect.Response[runtimev1.ListTopicSubscriptionsResponse], error) {
	subs := make([]*runtimev1.TopicSubscription, 0)
	for _, v := range s.topicRegistrar {
		s := v.Subscription
		sub := &runtimev1.TopicSubscription{
			PubsubName:      s.PubsubName,
			Topic:           s.Topic,
			Metadata:        s.Metadata,
			Routes:          convertRoutes(s.Routes),
			DeadLetterTopic: s.DeadLetterTopic,
		}
		subs = append(subs, sub)
	}

	return connect.NewResponse(&runtimev1.ListTopicSubscriptionsResponse{
		Subscriptions: subs,
	}), nil
}

func convertRoutes(routes *internal.TopicRoutes) *runtimev1.TopicRoutes {
	if routes == nil {
		return nil
	}
	rules := make([]*runtimev1.TopicRule, len(routes.Rules))
	for i, rule := range routes.Rules {
		rules[i] = &runtimev1.TopicRule{
			Match: rule.Match,
			Path:  rule.Path,
		}
	}
	return &runtimev1.TopicRoutes{
		Rules:   rules,
		Default: routes.Default,
	}
}

// OnTopicEvent fired whenever a message has been published to a topic that has been subscribed.
// Dapr sends published messages in a CloudEvents v1.0 envelope.
func (s *Server) OnTopicEvent(ctx context.Context, in *connect.Request[runtimev1.TopicEventRequest]) (*connect.Response[runtimev1.TopicEventResponse], error) {
	if in == nil || in.Msg.GetTopic() == "" || in.Msg.GetPubsubName() == "" {
		// this is really Dapr issue more than the event request format.
		// since Dapr will not get updated until long after this event expires, just drop it
		return connect.NewResponse(&runtimev1.TopicEventResponse{Status: runtimev1.TopicEventResponse_DROP}), errors.New("pub/sub and topic names required")
	}
	key := in.Msg.GetPubsubName() + "-" + in.Msg.GetTopic()
	noValidationKey := in.Msg.GetPubsubName()

	var sub *internal.TopicRegistration
	var ok bool

	sub, ok = s.topicRegistrar[key]
	if !ok {
		sub, ok = s.topicRegistrar[noValidationKey]
	}

	if ok {
		data := interface{}(in.Msg.GetData())
		if len(in.Msg.GetData()) > 0 {
			mediaType, _, err := mime.ParseMediaType(in.Msg.GetDataContentType())
			if err == nil {
				var v interface{}
				switch mediaType {
				case "application/json":
					if err := json.Unmarshal(in.Msg.GetData(), &v); err == nil {
						data = v
					}
				case "text/plain":
					// Assume UTF-8 encoded string.
					data = string(in.Msg.GetData())
				default:
					if strings.HasPrefix(mediaType, "application/") &&
						strings.HasSuffix(mediaType, "+json") {
						if err := json.Unmarshal(in.Msg.GetData(), &v); err == nil {
							data = v
						}
					}
				}
			}
		}

		e := &common.TopicEvent{
			ID:              in.Msg.GetId(),
			Source:          in.Msg.GetSource(),
			Type:            in.Msg.GetType(),
			SpecVersion:     in.Msg.GetSpecVersion(),
			DataContentType: in.Msg.GetDataContentType(),
			Data:            data,
			RawData:         in.Msg.GetData(),
			Topic:           in.Msg.GetTopic(),
			PubsubName:      in.Msg.GetPubsubName(),
			Metadata:        getCustomMetadataFromContext(ctx),
		}
		h := sub.DefaultHandler
		if in.Msg.GetPath() != "" {
			if pathHandler, ok := sub.RouteHandlers[in.Msg.GetPath()]; ok {
				h = pathHandler
			}
		}
		if h == nil {
			return connect.NewResponse(&runtimev1.TopicEventResponse{Status: runtimev1.TopicEventResponse_RETRY}), fmt.Errorf(
				"route %s for pub/sub and topic combination not configured: %s/%s",
				in.Msg.GetPath(), in.Msg.GetPubsubName(), in.Msg.GetTopic(),
			)
		}
		retry, err := h.Handle(ctx, e)
		if err == nil {
			return connect.NewResponse(&runtimev1.TopicEventResponse{Status: runtimev1.TopicEventResponse_SUCCESS}), nil
		}
		if retry {
			return connect.NewResponse(&runtimev1.TopicEventResponse{Status: runtimev1.TopicEventResponse_RETRY}), err
		}
		return connect.NewResponse(&runtimev1.TopicEventResponse{Status: runtimev1.TopicEventResponse_DROP}), nil
	}
	return connect.NewResponse(&runtimev1.TopicEventResponse{Status: runtimev1.TopicEventResponse_RETRY}), fmt.Errorf(
		"pub/sub and topic combination not configured: %s/%s",
		in.Msg.GetPubsubName(), in.Msg.GetTopic(),
	)
}

func getCustomMetadataFromContext(ctx context.Context) map[string]string {
	md := make(map[string]string)
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range meta {
			if strings.HasPrefix(strings.ToLower(k), "metadata.") {
				md[k[9:]] = v[0]
			}
		}
	}
	return md
}

func (s *Server) OnBulkTopicEventAlpha1(ctx context.Context, in *connect.Request[runtimev1.TopicEventBulkRequest]) (*connect.Response[runtimev1.TopicEventBulkResponse], error) {
	panic("This API callback is not supported.")
}
