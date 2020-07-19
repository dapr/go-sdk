package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/dapr/go-sdk/service"
	"github.com/stretchr/testify/assert"
)

func eventHandler(ctx context.Context, event *service.TopicEvent) error {
	if event == nil {
		return errors.New("nil event")
	}
	return nil
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestTopic$
func TestTopic(t *testing.T) {
	t.Parallel()

	topicName := "test"
	ctx := context.Background()

	server := getTestServer()
	server.AddTopicEventHandler(topicName, eventHandler)
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
			DataContentType: "text/plain",
			Source:          "test",
			SpecVersion:     "v0.3",
			Topic:           topicName,
			Type:            "test",
		}
		_, err := server.OnTopicEvent(ctx, in)
		assert.NoError(t, err)
	})

	stopTestServer(t, server)
}
