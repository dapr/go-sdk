package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/stretchr/testify/assert"
)

// go test -v -count=1 -run TestTopic ./server/grpc
func TestTopic(t *testing.T) {
	t.Parallel()

	topicName := "test"
	eventID := "1"
	dataContentType := "text/plain"

	server := getTestServer()
	server.AddTopicEventHandler(topicName, eventHandler)
	startTestServer(server)

	ctx := context.Background()
	in := &runtime.TopicEventRequest{
		Id:              eventID,
		DataContentType: dataContentType,
		Source:          "test",
		SpecVersion:     "v0.3",
		Topic:           topicName,
		Type:            "test",
	}

	_, err := server.OnTopicEvent(ctx, in)
	assert.NoError(t, err)

	stopTestServer(t, server)
}

func eventHandler(ctx context.Context, event *TopicEvent) error {
	if event == nil {
		return errors.New("nil event")
	}
	return nil
}
