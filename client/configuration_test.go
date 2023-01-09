package client

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	valueSuffix = "_value"
)

func TestGetConfigurationItem(t *testing.T) {
	ctx := context.Background()

	t.Run("get configuration item", func(t *testing.T) {
		resp, err := testClient.GetConfigurationItem(ctx, "example-config", "mykey")
		assert.Nil(t, err)
		assert.Equal(t, "mykey"+valueSuffix, resp.Value)
	})

	t.Run("get configuration item with invalid storeName", func(t *testing.T) {
		_, err := testClient.GetConfigurationItem(ctx, "", "mykey")
		assert.NotNil(t, err)
	})
}

func TestGetConfigurationItems(t *testing.T) {
	ctx := context.Background()

	keys := []string{"mykey1", "mykey2", "mykey3"}
	t.Run("Test get configuration items", func(t *testing.T) {
		resp, err := testClient.GetConfigurationItems(ctx, "example-config", keys)
		assert.Nil(t, err)
		for _, k := range keys {
			assert.Equal(t, k+valueSuffix, resp[k].Value)
		}
	})
}

func TestSubscribeConfigurationItems(t *testing.T) {
	ctx := context.Background()

	counter := 0
	totalCounter := 0
	keys := []string{"mykey1", "mykey2", "mykey3"}
	t.Run("Test subscribe configuration items", func(t *testing.T) {
		err := testClient.SubscribeConfigurationItems(ctx, "example-config",
			keys, func(s string, items map[string]*ConfigurationItem) {
				counter++
				for _, k := range keys {
					assert.Equal(t, k+valueSuffix, items[k].Value)
					totalCounter++
				}
			})
		assert.Nil(t, err)
	})
	time.Sleep(time.Second*5 + time.Millisecond*500)
	assert.Equal(t, 5, counter)
	assert.Equal(t, 15, totalCounter)
}

func TestUnSubscribeConfigurationItems(t *testing.T) {
	ctx := context.Background()

	var counter, totalCounter uint32
	t.Run("Test unsubscribe configuration items", func(t *testing.T) {
		subscribeIDChan := make(chan string)
		go func() {
			keys := []string{"mykey1", "mykey2", "mykey3"}
			err := testClient.SubscribeConfigurationItems(ctx, "example-config",
				keys, func(id string, items map[string]*ConfigurationItem) {
					atomic.AddUint32(&counter, 1)
					for _, k := range keys {
						assert.Equal(t, k+valueSuffix, items[k].Value)
						atomic.AddUint32(&totalCounter, 1)
					}
					select {
					case subscribeIDChan <- id:
					default:
					}
				})
			assert.Nil(t, err)
		}()
		subscribeID := <-subscribeIDChan
		time.Sleep(time.Second * 2)
		time.Sleep(time.Millisecond * 500)
		err := testClient.UnsubscribeConfigurationItems(ctx, "example-config", subscribeID)
		assert.Nil(t, err)
	})
	time.Sleep(time.Second * 5)
	assert.Equal(t, uint32(3), atomic.LoadUint32(&counter))
	assert.Equal(t, uint32(9), atomic.LoadUint32(&totalCounter))
}
