package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/stretchr/testify/assert"
)

func eventHandler(ctx context.Context, event *TopicEvent) error {
	if event == nil {
		return errors.New("nil event")
	}
	return nil
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestTopic$
func TestTopic(t *testing.T) {
	t.Parallel()

	topicName := "test"
	componentName := "messages"
	ctx := context.Background()

	server := getTestServer()
	err := server.AddTopicEventHandler(componentName, topicName, eventHandler)
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
			SpecVersion:     "v0.3",
			DataContentType: "text/plain",
			Data:            []byte("test"),
			Topic:           topicName,
			PubsubName:      componentName,
		}
		_, err := server.OnTopicEvent(ctx, in)
		assert.NoError(t, err)
	})

	stopTestServer(t, server)
}
