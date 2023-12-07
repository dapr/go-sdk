package internal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

func TestTopicRegistrarValidation(t *testing.T) {
	fn := func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		return false, nil
	}
	tests := map[string]struct {
		sub common.Subscription
		fn  common.TopicEventHandler
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
	t.Run("with subscription", func(t *testing.T) {
		for name, tt := range tests {
			tt := tt // dereference loop var
			t.Run(name, func(t *testing.T) {
				m := internal.TopicRegistrar{}
				if tt.err != "" {
					assert.EqualError(t, m.AddSubscription(&tt.sub, tests[name].fn), tt.err)
				} else {
					assert.NoError(t, m.AddSubscription(&tt.sub, tt.fn))
				}
			})
		}
	})
	t.Run("with bulk subscription", func(t *testing.T) {
		for name, tt := range tests {
			tt := tt // dereference loop var
			t.Run(name, func(t *testing.T) {
				m := internal.TopicRegistrar{}
				if tt.err != "" {
					assert.EqualError(t, m.AddBulkSubscription(&tt.sub, tests[name].fn, 10, 1000), tt.err)
				} else {
					assert.NoError(t, m.AddBulkSubscription(&tt.sub, tt.fn, 10, 1000))
				}
			})
		}
	})
}

func TestTopicAddSubscriptionMetadata(t *testing.T) {
	handler := func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		return false, nil
	}
	sub := &common.Subscription{
		PubsubName: "pubsubname",
		Topic:      "topic",
		Metadata:   map[string]string{"key": "value"},
	}

	t.Run("with subscription", func(t *testing.T) {
		topicRegistrar := internal.TopicRegistrar{}
		assert.NoError(t, topicRegistrar.AddSubscription(sub, handler))

		actual := topicRegistrar["pubsubname-topic"].Subscription
		expected := &internal.TopicSubscription{
			PubsubName: sub.PubsubName,
			Topic:      sub.Topic,
			Metadata:   sub.Metadata,
		}
		assert.Equal(t, expected, actual)
	})
	t.Run("with bulk subscription", func(t *testing.T) {
		topicRegistrar := internal.TopicRegistrar{}
		assert.NoError(t, topicRegistrar.AddBulkSubscription(sub, handler, 10, 1000))

		actual := topicRegistrar["pubsubname-topic"].Subscription
		expected := &internal.TopicSubscription{
			PubsubName: sub.PubsubName,
			Topic:      sub.Topic,
			Metadata:   sub.Metadata,
			BulkSubscribe: &internal.BulkSubscribeOptions{
				Enabled:            true,
				MaxMessagesCount:   10,
				MaxAwaitDurationMs: 1000,
			},
		}
		assert.Equal(t, expected, actual)
	})
}
