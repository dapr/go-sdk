package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
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
	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
	}
	server := getTestServer()
	err := server.AddTopicEventHandler(sub, eventHandler)
	assert.Nil(t, err)
	resp, err := server.ListTopicSubscriptions(context.Background(), &empty.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Lenf(t, resp.Subscriptions, 1, "expected 1 handlers")
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
			Topic: "invlid",
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
		assert.Error(t, err)
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
