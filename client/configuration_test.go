package client

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const (
	valueSuffix = "_value"
)

func TestGetConfigurationItem(t *testing.T) {
	ctx := t.Context()

	t.Run("get configuration item", func(t *testing.T) {
		resp, err := testClient.GetConfigurationItem(ctx, "example-config", "mykey")
		require.NoError(t, err)
		assert.Equal(t, "mykey"+valueSuffix, resp.Value)
	})

	t.Run("get configuration item with invalid storeName", func(t *testing.T) {
		_, err := testClient.GetConfigurationItem(ctx, "", "mykey")
		require.Error(t, err)
	})
}

func TestGetConfigurationItems(t *testing.T) {
	ctx := t.Context()

	keys := []string{"mykey1", "mykey2", "mykey3"}
	t.Run("Test get configuration items", func(t *testing.T) {
		resp, err := testClient.GetConfigurationItems(ctx, "example-config", keys)
		require.NoError(t, err)
		for _, k := range keys {
			assert.Equal(t, k+valueSuffix, resp[k].Value)
		}
	})
}

func TestSubscribeConfigurationItems(t *testing.T) {
	ctx := t.Context()

	var counter, totalCounter uint32
	counter = 0
	totalCounter = 0
	keys := []string{"mykey1", "mykey2", "mykey3"}
	t.Run("Test subscribe configuration items", func(t *testing.T) {
		_, err := testClient.SubscribeConfigurationItems(ctx, "example-config",
			keys, func(s string, items map[string]*ConfigurationItem) {
				atomic.AddUint32(&counter, 1)
				for _, k := range keys {
					assert.Equal(t, k+valueSuffix, items[k].Value)
					atomic.AddUint32(&totalCounter, 1)
				}
			})
		require.NoError(t, err)
	})
	time.Sleep(time.Second*5 + time.Millisecond*500)
	assert.Equal(t, uint32(5), atomic.LoadUint32(&counter))
	assert.Equal(t, uint32(15), atomic.LoadUint32(&totalCounter))
}

func TestUnSubscribeConfigurationItems(t *testing.T) {
	ctx := t.Context()

	var counter, totalCounter uint32
	t.Run("Test unsubscribe configuration items", func(t *testing.T) {
		keys := []string{"mykey1", "mykey2", "mykey3"}
		subscribeID, err := testClient.SubscribeConfigurationItems(ctx, "example-config",
			keys, func(id string, items map[string]*ConfigurationItem) {
				atomic.AddUint32(&counter, 1)
				for _, k := range keys {
					assert.Equal(t, k+valueSuffix, items[k].Value)
					atomic.AddUint32(&totalCounter, 1)
				}
			})
		require.NoError(t, err)
		time.Sleep(time.Second * 2)
		time.Sleep(time.Millisecond * 500)
		err = testClient.UnsubscribeConfigurationItems(ctx, "example-config", subscribeID)
		require.NoError(t, err)
	})
	time.Sleep(time.Second * 5)
	assert.Equal(t, uint32(3), atomic.LoadUint32(&counter))
	assert.Equal(t, uint32(9), atomic.LoadUint32(&totalCounter))
}
