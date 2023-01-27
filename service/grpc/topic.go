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

package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"

	runtimev1pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// AddTopicEventHandler appends provided event handler with topic name to the service.
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn common.TopicEventHandler) error {
	if sub == nil {
		return errors.New("subscription required")
	}
	if err := s.topicRegistrar.AddSubscription(sub, fn); err != nil {
		return err
	}

	return nil
}

// ListTopicSubscriptions is called by Dapr to get the list of topics in a pubsub component the app wants to subscribe to.
func (s *Server) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*runtimev1pb.ListTopicSubscriptionsResponse, error) {
	subs := make([]*runtimev1pb.TopicSubscription, 0)
	for _, v := range s.topicRegistrar {
		s := v.Subscription
		sub := &runtimev1pb.TopicSubscription{
			PubsubName: s.PubsubName,
			Topic:      s.Topic,
			Metadata:   s.Metadata,
			Routes:     convertRoutes(s.Routes),
		}
		subs = append(subs, sub)
	}

	return &runtimev1pb.ListTopicSubscriptionsResponse{
		Subscriptions: subs,
	}, nil
}

func convertRoutes(routes *internal.TopicRoutes) *runtimev1pb.TopicRoutes {
	if routes == nil {
		return nil
	}
	rules := make([]*runtimev1pb.TopicRule, len(routes.Rules))
	for i, rule := range routes.Rules {
		rules[i] = &runtimev1pb.TopicRule{
			Match: rule.Match,
			Path:  rule.Path,
		}
	}
	return &runtimev1pb.TopicRoutes{
		Rules:   rules,
		Default: routes.Default,
	}
}

// OnTopicEvent fired whenever a message has been published to a topic that has been subscribed.
// Dapr sends published messages in a CloudEvents v1.0 envelope.
func (s *Server) OnTopicEvent(ctx context.Context, in *runtimev1pb.TopicEventRequest) (*runtimev1pb.TopicEventResponse, error) {
	if in == nil || in.Topic == "" || in.PubsubName == "" {
		// this is really Dapr issue more than the event request format.
		// since Dapr will not get updated until long after this event expires, just drop it
		return &runtimev1pb.TopicEventResponse{Status: runtimev1pb.TopicEventResponse_DROP}, errors.New("pub/sub and topic names required")
	}
	key := in.PubsubName + "-" + in.Topic
	noValidationKey := in.PubsubName

	var sub *internal.TopicRegistration
	var ok bool

	sub, ok = s.topicRegistrar[key]
	if !ok {
		sub, ok = s.topicRegistrar[noValidationKey]
	}

	if ok {
		data := interface{}(in.Data)
		if len(in.Data) > 0 {
			mediaType, _, err := mime.ParseMediaType(in.DataContentType)
			if err == nil {
				var v interface{}
				switch mediaType {
				case "application/json":
					if err := json.Unmarshal(in.Data, &v); err == nil {
						data = v
					}
				case "text/plain":
					// Assume UTF-8 encoded string.
					data = string(in.Data)
				default:
					if strings.HasPrefix(mediaType, "application/") &&
						strings.HasSuffix(mediaType, "+json") {
						if err := json.Unmarshal(in.Data, &v); err == nil {
							data = v
						}
					}
				}
			}
		}

		e := &common.TopicEvent{
			ID:              in.Id,
			Source:          in.Source,
			Type:            in.Type,
			SpecVersion:     in.SpecVersion,
			DataContentType: in.DataContentType,
			Data:            data,
			RawData:         in.Data,
			Topic:           in.Topic,
			PubsubName:      in.PubsubName,
		}
		h := sub.DefaultHandler
		if in.Path != "" {
			if pathHandler, ok := sub.RouteHandlers[in.Path]; ok {
				h = pathHandler
			}
		}
		if h == nil {
			return &runtimev1pb.TopicEventResponse{Status: runtimev1pb.TopicEventResponse_RETRY}, fmt.Errorf(
				"route %s for pub/sub and topic combination not configured: %s/%s",
				in.Path, in.PubsubName, in.Topic,
			)
		}
		retry, err := h(ctx, e)
		if err == nil {
			return &runtimev1pb.TopicEventResponse{Status: runtimev1pb.TopicEventResponse_SUCCESS}, nil
		}
		if retry {
			return &runtimev1pb.TopicEventResponse{Status: runtimev1pb.TopicEventResponse_RETRY}, err
		}
		return &runtimev1pb.TopicEventResponse{Status: runtimev1pb.TopicEventResponse_DROP}, nil
	}
	return &runtimev1pb.TopicEventResponse{Status: runtimev1pb.TopicEventResponse_RETRY}, fmt.Errorf(
		"pub/sub and topic combination not configured: %s/%s",
		in.PubsubName, in.Topic,
	)
}
