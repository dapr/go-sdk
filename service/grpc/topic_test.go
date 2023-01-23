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
	"errors"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"

	runtime "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
)

func TestTopicErrors(t *testing.T) {
	server := getTestServer()
	err := server.AddTopicEventHandler(nil, nil)
	assert.Errorf(t, err, "expected error on nil sub")

	sub := &common.Subscription{}
	err = server.AddTopicEventHandler(sub, nil)
	assert.Errorf(t, err, "expected error on invalid sub")

	sub.PubsubName = "messages"
	err = server.AddTopicEventHandler(sub, nil)
	assert.Errorf(t, err, "expected error on sub without topic")

	sub.Topic = "test"
	err = server.AddTopicEventHandler(sub, nil)
	assert.Errorf(t, err, "expected error on sub without handler")
}

func TestTopicSubscriptionList(t *testing.T) {
	server := getTestServer()

	// Add default route.
	sub1 := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
		Route:      "/test",
	}
	err := server.AddTopicEventHandler(sub1, eventHandler)
	assert.Nil(t, err)
	resp, err := server.ListTopicSubscriptions(context.Background(), &empty.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	if assert.Lenf(t, resp.Subscriptions, 1, "expected 1 handlers") {
		sub := resp.Subscriptions[0]
		assert.Equal(t, "messages", sub.PubsubName)
		assert.Equal(t, "test", sub.Topic)
		assert.Nil(t, sub.Routes)
	}

	// Add routing rule.
	sub2 := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
		Route:      "/other",
		Match:      `event.type == "other"`,
	}
	err = server.AddTopicEventHandler(sub2, eventHandler)
	assert.Nil(t, err)
	resp, err = server.ListTopicSubscriptions(context.Background(), &empty.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	if assert.Lenf(t, resp.Subscriptions, 1, "expected 1 handlers") {
		sub := resp.Subscriptions[0]
		assert.Equal(t, "messages", sub.PubsubName)
		assert.Equal(t, "test", sub.Topic)
		if assert.NotNil(t, sub.Routes) {
			assert.Equal(t, "/test", sub.Routes.Default)
			if assert.Len(t, sub.Routes.Rules, 1) {
				rule := sub.Routes.Rules[0]
				assert.Equal(t, "/other", rule.Path)
				assert.Equal(t, `event.type == "other"`, rule.Match)
			}
		}
	}
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestTopic$
func TestTopic(t *testing.T) {
	ctx := context.Background()

	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
	}
	server := getTestServer()

	err := server.AddTopicEventHandler(sub, eventHandler)
	assert.Nil(t, err)

	startTestServer(server)

	t.Run("topic event without request", func(t *testing.T) {
		_, err := server.OnTopicEvent(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("topic event for wrong topic", func(t *testing.T) {
		in := &runtime.TopicEventRequest{
			Topic: "invalid",
		}
		_, err := server.OnTopicEvent(ctx, in)
		assert.Error(t, err)
	})

	t.Run("topic event for valid topic", func(t *testing.T) {
		in := &runtime.TopicEventRequest{
			Id:              "a123",
			Source:          "test",
			Type:            "test",
			SpecVersion:     "v1.0",
			DataContentType: "text/plain",
			Data:            []byte("test"),
			Topic:           sub.Topic,
			PubsubName:      sub.PubsubName,
		}
		_, err := server.OnTopicEvent(ctx, in)
		assert.NoError(t, err)
	})

	stopTestServer(t, server)
}

func TestTopicWithValidationDisabled(t *testing.T) {
	ctx := context.Background()

	sub := &common.Subscription{
		PubsubName:             "messages",
		Topic:                  "*",
		DisableTopicValidation: true,
	}
	server := getTestServer()

	err := server.AddTopicEventHandler(sub, eventHandler)
	assert.Nil(t, err)

	startTestServer(server)

	in := &runtime.TopicEventRequest{
		Id:              "a123",
		Source:          "test",
		Type:            "test",
		SpecVersion:     "v1.0",
		DataContentType: "text/plain",
		Data:            []byte("test"),
		Topic:           "test",
		PubsubName:      sub.PubsubName,
	}

	_, err = server.OnTopicEvent(ctx, in)
	assert.NoError(t, err)
}

func TestTopicWithErrors(t *testing.T) {
	ctx := context.Background()

	sub1 := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test1",
	}

	sub2 := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test2",
	}
	server := getTestServer()

	err := server.AddTopicEventHandler(sub1, eventHandlerWithRetryError)
	assert.Nil(t, err)

	err = server.AddTopicEventHandler(sub2, eventHandlerWithError)
	assert.Nil(t, err)

	startTestServer(server)

	t.Run("topic event for retry error", func(t *testing.T) {
		in := &runtime.TopicEventRequest{
			Id:              "a123",
			Source:          "test",
			Type:            "test",
			SpecVersion:     "v1.0",
			DataContentType: "text/plain",
			Data:            []byte("test"),
			Topic:           sub1.Topic,
			PubsubName:      sub1.PubsubName,
		}
		resp, err := server.OnTopicEvent(ctx, in)
		assert.Error(t, err)
		assert.Equal(t, resp.GetStatus(), runtime.TopicEventResponse_RETRY)
	})

	t.Run("topic event for error", func(t *testing.T) {
		in := &runtime.TopicEventRequest{
			Id:              "a123",
			Source:          "test",
			Type:            "test",
			SpecVersion:     "v1.0",
			DataContentType: "text/plain",
			Data:            []byte("test"),
			Topic:           sub2.Topic,
			PubsubName:      sub2.PubsubName,
		}
		resp, err := server.OnTopicEvent(ctx, in)
		assert.NoError(t, err)
		assert.Equal(t, resp.GetStatus(), runtime.TopicEventResponse_DROP)
	})

	stopTestServer(t, server)
}

func eventHandler(ctx context.Context, event *common.TopicEvent) (retry bool, err error) {
	if event == nil {
		return true, errors.New("nil event")
	}
	return false, nil
}

func eventHandlerWithRetryError(ctx context.Context, event *common.TopicEvent) (retry bool, err error) {
	return true, errors.New("nil event")
}

func eventHandlerWithError(ctx context.Context, event *common.TopicEvent) (retry bool, err error) {
	return false, errors.New("nil event")
}

func TestEventDataHandling(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		contentType string
		data        string
		value       interface{}
	}{
		"JSON bytes": {
			contentType: "application/json",
			data:        `{"message":"hello"}`,
			value: map[string]interface{}{
				"message": "hello",
			},
		},
		"JSON entension media type bytes": {
			contentType: "application/extension+json",
			data:        `{"message":"hello"}`,
			value: map[string]interface{}{
				"message": "hello",
			},
		},
		"Test": {
			contentType: "text/plain",
			data:        `message = hello`,
			value:       `message = hello`,
		},
		"Other": {
			contentType: "application/octet-stream",
			data:        `message = hello`,
			value:       []byte(`message = hello`),
		},
	}

	s := getTestServer()

	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
		Route:      "/test",
		Metadata:   map[string]string{},
	}

	recv := make(chan struct{}, 1)
	var topicEvent *common.TopicEvent
	handler := func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		topicEvent = e
		recv <- struct{}{}

		return false, nil
	}
	err := s.AddTopicEventHandler(sub, handler)
	assert.NoErrorf(t, err, "error adding event handler")

	startTestServer(s)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			in := runtime.TopicEventRequest{
				Id:              "a123",
				Source:          "test",
				Type:            "test",
				SpecVersion:     "v1.0",
				DataContentType: tt.contentType,
				Data:            []byte(tt.data),
				Topic:           sub.Topic,
				PubsubName:      sub.PubsubName,
			}

			s.OnTopicEvent(ctx, &in)
			<-recv
			assert.Equal(t, tt.value, topicEvent.Data)
		})
	}
}
