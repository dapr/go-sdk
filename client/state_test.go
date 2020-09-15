package client

import (
	"context"
	"testing"
	"time"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	"github.com/stretchr/testify/assert"
)

func TestDurationConverter(t *testing.T) {
	d := time.Duration(10 * time.Second)
	pd := toProtoDuration(d)
	assert.NotNil(t, pd)
	assert.Equal(t, pd.Seconds, int64(10))
}

func TestStateOptionsConverter(t *testing.T) {
	s := &StateOptions{
		Concurrency: StateConcurrencyLastWrite,
		Consistency: StateConsistencyStrong,
	}
	p := toProtoStateOptions(s)
	assert.NotNil(t, p)
	assert.Equal(t, p.Concurrency, v1.StateOptions_CONCURRENCY_LAST_WRITE)
	assert.Equal(t, p.Consistency, v1.StateOptions_CONSISTENCY_STRONG)
}

// go test -timeout 30s ./client -count 1 -run ^TestSaveState$
func TestSaveState(t *testing.T) {
	ctx := context.Background()
	data := "test"
	store := "test"
	key := "key1"

	t.Run("save data", func(t *testing.T) {
		err := testClient.SaveState(ctx, store, key, []byte(data))
		assert.Nil(t, err)
	})

	t.Run("get saved data", func(t *testing.T) {
		item, err := testClient.GetState(ctx, store, key)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.NotEmpty(t, item.Etag)
		assert.Equal(t, item.Key, key)
		assert.Equal(t, string(item.Value), data)
	})

	t.Run("get saved data with consistency", func(t *testing.T) {
		item, err := testClient.GetStateWithConsistency(ctx, store, key, nil, StateConsistencyStrong)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.NotEmpty(t, item.Etag)
		assert.Equal(t, item.Key, key)
		assert.Equal(t, string(item.Value), data)
	})

	t.Run("save data with version", func(t *testing.T) {
		item := &SetStateItem{
			Etag:  "1",
			Key:   key,
			Value: []byte(data),
		}
		err := testClient.SaveStateItems(ctx, store, item)
		assert.Nil(t, err)
	})

	t.Run("delete data", func(t *testing.T) {
		err := testClient.DeleteState(ctx, store, key)
		assert.Nil(t, err)
	})
}

// go test -timeout 30s ./client -count 1 -run ^TestStateTransactions$
func TestStateTransactions(t *testing.T) {
	ctx := context.Background()
	data := `{ "message": "test" }`
	store := "test"
	meta := map[string]string{}
	keys := []string{"k1", "k2", "k3"}
	adds := make([]*StateOperation, 0)

	for _, k := range keys {
		op := &StateOperation{
			Type: StateOperationTypeUpsert,
			Item: &SetStateItem{
				Key:   k,
				Value: []byte(data),
			},
		}
		adds = append(adds, op)
	}

	t.Run("exec inserts", func(t *testing.T) {
		err := testClient.ExecuteStateTransaction(ctx, store, meta, adds)
		assert.Nil(t, err)
	})

	t.Run("exec upserts", func(t *testing.T) {
		items, err := testClient.GetBulkItems(ctx, store, keys, 10)
		assert.Nil(t, err)
		assert.NotNil(t, items)
		assert.Len(t, items, len(keys))

		upsers := make([]*StateOperation, 0)
		for _, item := range items {
			op := &StateOperation{
				Type: StateOperationTypeUpsert,
				Item: &SetStateItem{
					Key:   item.Key,
					Etag:  item.Etag,
					Value: item.Value,
				},
			}
			upsers = append(upsers, op)
		}
		err = testClient.ExecuteStateTransaction(ctx, store, meta, upsers)
		assert.Nil(t, err)
	})

	t.Run("get and validate inserts", func(t *testing.T) {
		items, err := testClient.GetBulkItems(ctx, store, keys, 10)
		assert.Nil(t, err)
		assert.NotNil(t, items)
		assert.Len(t, items, len(keys))
		assert.Equal(t, data, string(items[0].Value))
	})

	for _, op := range adds {
		op.Type = StateOperationTypeDelete
	}

	t.Run("exec deletes", func(t *testing.T) {
		err := testClient.ExecuteStateTransaction(ctx, store, meta, adds)
		assert.Nil(t, err)
	})

	t.Run("ensure deletes", func(t *testing.T) {
		items, err := testClient.GetBulkItems(ctx, store, keys, 3)
		assert.Nil(t, err)
		assert.NotNil(t, items)
		assert.Len(t, items, 0)
	})

}
