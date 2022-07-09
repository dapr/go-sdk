package client

import (
	"context"
	"strconv"
	"testing"
	"time"

	"go.uber.org/atomic"

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

	t.Run("Test get configuration items", func(t *testing.T) {
		resp, err := testClient.GetConfigurationItems(ctx, "example-config", []string{"mykey1", "mykey2", "mykey3"})
		assert.Nil(t, err)
		for k, v := range resp {
			assert.Equal(t, k+valueSuffix, v.Value)
		}
	})
}

func TestSubscribeConfigurationItems(t *testing.T) {
	ctx := context.Background()

	counter := 0
	totalCounter := 0
	t.Run("Test subscribe configuration items", func(t *testing.T) {
		err := testClient.SubscribeConfigurationItems(ctx, "example-config",
			[]string{"mykey", "mykey2", "mykey3"}, func(s string, items map[string]*ConfigurationItem) {
				counter++
				for k, v := range items {
					assert.Equal(t, v.Value, k+"_"+strconv.Itoa(counter-1))
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

	counter := atomic.Int32{}
	totalCounter := atomic.Int32{}
	t.Run("Test unsubscribe configuration items", func(t *testing.T) {
		subscribeID := ""
		subscribeIDChan := make(chan string)
		go func() {
			err := testClient.SubscribeConfigurationItems(ctx, "example-config",
				[]string{"mykey", "mykey2", "mykey3"}, func(id string, items map[string]*ConfigurationItem) {
					counter.Inc()
					for k, v := range items {
						assert.Equal(t, v.Value, k+"_"+strconv.Itoa(int(counter.Load()-1)))
						totalCounter.Inc()
					}
					select {
					case subscribeIDChan <- id:
					default:
					}
				})
			assert.Nil(t, err)
		}()
		subscribeID = <-subscribeIDChan
		time.Sleep(time.Second * 2)
		time.Sleep(time.Millisecond * 500)
		err := testClient.UnsubscribeConfigurationItems(ctx, "example-config", subscribeID)
		assert.Nil(t, err)
	})
	time.Sleep(time.Second * 5)
	assert.Equal(t, 3, int(counter.Load()))
	assert.Equal(t, 9, int(totalCounter.Load()))
}
