package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dapr/go-sdk/service/internal"
)

func TestTopicSubscripiton(t *testing.T) {
	t.Run("duplicate metadata", func(t *testing.T) {
		sub := internal.NewTopicSubscription("test", "mytopic")
		assert.NoError(t, sub.SetMetadata(map[string]string{
			"test": "test",
		}))
		assert.EqualError(t, sub.SetMetadata(map[string]string{
			"test": "test",
		}), "subscription for topic mytopic on pubsub test already has metadata set")
	})

	t.Run("duplicate route", func(t *testing.T) {
		sub := internal.NewTopicSubscription("test", "mytopic")
		assert.NoError(t, sub.SetDefaultRoute("/test"))
		assert.Equal(t, "/test", sub.Route)
		assert.EqualError(t, sub.SetDefaultRoute("/test"),
			"subscription for topic mytopic on pubsub test already has route /test")
	})

	t.Run("duplicate route after routing rule", func(t *testing.T) {
		sub := internal.NewTopicSubscription("test", "mytopic")
		assert.NoError(t, sub.AddRoutingRule("/other", `event.type == "test"`, 0))
		assert.NoError(t, sub.SetDefaultRoute("/test"))
		assert.EqualError(t, sub.SetDefaultRoute("/test"),
			"subscription for topic mytopic on pubsub test already has route /test")
	})

	t.Run("default route after routing rule", func(t *testing.T) {
		sub := internal.NewTopicSubscription("test", "mytopic")
		assert.NoError(t, sub.SetDefaultRoute("/test"))
		assert.Equal(t, "/test", sub.Route)
		assert.NoError(t, sub.AddRoutingRule("/other", `event.type == "test"`, 0))
		assert.Equal(t, "", sub.Route)
		assert.Equal(t, "/test", sub.Routes.Default)
		assert.EqualError(t, sub.SetDefaultRoute("/test"),
			"subscription for topic mytopic on pubsub test already has route /test")
	})

	t.Run("duplicate routing rule priority", func(t *testing.T) {
		sub := internal.NewTopicSubscription("test", "mytopic")
		assert.NoError(t, sub.AddRoutingRule("/other", `event.type == "other"`, 1))
		assert.EqualError(t, sub.AddRoutingRule("/test", `event.type == "test"`, 1),
			"subscription for topic mytopic on pubsub test already has a routing rule with priority 1")
	})

	t.Run("priority ordering", func(t *testing.T) {
		sub := internal.NewTopicSubscription("test", "mytopic")
		assert.NoError(t, sub.AddRoutingRule("/100", `event.type == "100"`, 100))
		assert.NoError(t, sub.AddRoutingRule("/1", `event.type == "1"`, 1))
		assert.NoError(t, sub.AddRoutingRule("/50", `event.type == "50"`, 50))
		assert.NoError(t, sub.SetDefaultRoute("/default"))
		assert.Equal(t, "/default", sub.Routes.Default)
		if assert.Len(t, sub.Routes.Rules, 3) {
			assert.Equal(t, "/1", sub.Routes.Rules[0].Path)
			assert.Equal(t, `event.type == "1"`, sub.Routes.Rules[0].Match)
			assert.Equal(t, "/50", sub.Routes.Rules[1].Path)
			assert.Equal(t, `event.type == "50"`, sub.Routes.Rules[1].Match)
			assert.Equal(t, "/100", sub.Routes.Rules[2].Path)
			assert.Equal(t, `event.type == "100"`, sub.Routes.Rules[2].Match)
		}
	})
}
