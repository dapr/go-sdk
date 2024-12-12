package internal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

func TestTopicRegistrarValidation(t *testing.T) {
	fn := common.TopicEventHandler(func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		return false, nil
	})
	tests := map[string]struct {
		sub common.Subscription
		fn  common.TopicEventSubscriber
		err string
	}{
		"pubsub required": {
			common.Subscription{ //nolint:exhaustivestruct
				PubsubName: "",
				Topic:      "test",
			}, fn, "pub/sub name required",
		},
		"topic required": {
			common.Subscription{ //nolint:exhaustivestruct
				PubsubName: "test",
				Topic:      "",
			}, fn, "topic name required",
		},
		"handler required": {
			common.Subscription{ //nolint:exhaustivestruct
				PubsubName: "test",
				Topic:      "test",
			}, nil, "topic handler required",
		},
		"route required for routing rule": {
			common.Subscription{ //nolint:exhaustivestruct
				PubsubName: "test",
				Topic:      "test",
				Route:      "",
				Match:      `event.type == "test"`,
			}, fn, "path is required for routing rules",
		},
		"success default route": {
			common.Subscription{ //nolint:exhaustivestruct
				PubsubName: "test",
				Topic:      "test",
			}, fn, "",
		},
		"success routing rule": {
			common.Subscription{ //nolint:exhaustivestruct
				PubsubName: "test",
				Topic:      "test",
				Route:      "/test",
				Match:      `event.type == "test"`,
			}, fn, "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := internal.TopicRegistrar{}
			if tt.err != "" {
				require.EqualError(t, m.AddSubscription(&tt.sub, tests[name].fn), tt.err)
			} else {
				require.NoError(t, m.AddSubscription(&tt.sub, tt.fn))
			}
		})
	}
}

func TestTopicAddSubscriptionMetadata(t *testing.T) {
	handler := common.TopicEventHandler(func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		return false, nil
	})
	topicRegistrar := internal.TopicRegistrar{}
	sub := &common.Subscription{
		PubsubName: "pubsubname",
		Topic:      "topic",
		Metadata:   map[string]string{"key": "value"},
	}

	require.NoError(t, topicRegistrar.AddSubscription(sub, handler))

	actual := topicRegistrar["pubsubname-topic"].Subscription
	expected := &internal.TopicSubscription{
		PubsubName: sub.PubsubName,
		Topic:      sub.Topic,
		Metadata:   sub.Metadata,
	}
	assert.Equal(t, expected, actual)
}
