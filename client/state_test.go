/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
)

const (
	testData  = "test"
	testStore = "store"
)

func TestTypes(t *testing.T) {
	var op OperationType = -1
	assert.Equal(t, UndefinedType, op.String())
	var c StateConcurrency = -1
	assert.Equal(t, UndefinedType, c.String())
	var d StateConsistency = -1
	assert.Equal(t, UndefinedType, d.String())
}

func TestDurationConverter(t *testing.T) {
	d := 10 * time.Second
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
	data := testData
	store := testStore
	key := "key1"

	t.Run("save data", func(t *testing.T) {
		err := testClient.SaveState(ctx, store, key, []byte(data), nil)
		assert.Nil(t, err)
	})

	t.Run("get saved data", func(t *testing.T) {
		item, err := testClient.GetState(ctx, store, key, nil)
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
		err := testClient.SaveStateWithETag(ctx, store, key, []byte(data), "1", nil)
		assert.Nil(t, err)
	})

	t.Run("delete data", func(t *testing.T) {
		err := testClient.DeleteState(ctx, store, key, nil)
		assert.Nil(t, err)
	})
}

// go test -timeout 30s ./client -count 1 -run ^TestDeleteState$
func TestDeleteState(t *testing.T) {
	ctx := context.Background()
	data := testData
	store := testStore
	key := "key1"

	t.Run("delete not exist data", func(t *testing.T) {
		err := testClient.DeleteState(ctx, store, key, nil)
		assert.Nil(t, err)
	})
	t.Run("delete not exist data with etag and meta", func(t *testing.T) {
		err := testClient.DeleteStateWithETag(ctx, store, key, &ETag{Value: "100"}, map[string]string{"meta1": "value1"},
			&StateOptions{Concurrency: StateConcurrencyFirstWrite, Consistency: StateConsistencyEventual})
		assert.Nil(t, err)
	})

	t.Run("save data", func(t *testing.T) {
		err := testClient.SaveState(ctx, store, key, []byte(data), nil)
		assert.Nil(t, err)
	})
	t.Run("confirm data saved", func(t *testing.T) {
		item, err := testClient.GetState(ctx, store, key, nil)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.NotEmpty(t, item.Etag)
		assert.Equal(t, item.Key, key)
		assert.Equal(t, string(item.Value), data)
	})

	t.Run("delete exist data", func(t *testing.T) {
		err := testClient.DeleteState(ctx, store, key, nil)
		assert.Nil(t, err)
	})
	t.Run("confirm data deleted", func(t *testing.T) {
		item, err := testClient.GetState(ctx, store, key, nil)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.NotEmpty(t, item.Etag)
		assert.Equal(t, item.Key, key)
		assert.Nil(t, item.Value)
	})

	t.Run("save data again with etag, meta", func(t *testing.T) {
		meta := map[string]string{"meta1": "value1"}
		err := testClient.SaveStateWithETag(ctx, store, key, []byte(data), "1", meta, WithConsistency(StateConsistencyEventual), WithConcurrency(StateConcurrencyFirstWrite))
		assert.Nil(t, err)
	})
	t.Run("confirm data saved", func(t *testing.T) {
		item, err := testClient.GetStateWithConsistency(ctx, store, key, map[string]string{"meta1": "value1"}, StateConsistencyEventual)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.NotEmpty(t, item.Etag)
		assert.Equal(t, item.Key, key)
		assert.Equal(t, string(item.Value), data)
	})

	t.Run("delete exist data with etag and meta", func(t *testing.T) {
		err := testClient.DeleteStateWithETag(ctx, store, key, &ETag{Value: "100"}, map[string]string{"meta1": "value1"},
			&StateOptions{Concurrency: StateConcurrencyFirstWrite, Consistency: StateConsistencyEventual})
		assert.Nil(t, err)
	})
	t.Run("confirm data deleted", func(t *testing.T) {
		item, err := testClient.GetStateWithConsistency(ctx, store, key, map[string]string{"meta1": "value1"}, StateConsistencyEventual)
		assert.Nil(t, err)
		assert.NotNil(t, item)
		assert.NotEmpty(t, item.Etag)
		assert.Equal(t, item.Key, key)
		assert.Nil(t, item.Value)
	})
}

func TestDeleteBulkState(t *testing.T) {
	ctx := context.Background()
	data := testData
	store := testStore
	keys := []string{"key1", "key2", "key3"}

	t.Run("delete not exist data", func(t *testing.T) {
		err := testClient.DeleteBulkState(ctx, store, keys, nil)
		assert.Nil(t, err)
	})

	t.Run("delete not exist data with stateIem", func(t *testing.T) {
		items := make([]*DeleteStateItem, 0, len(keys))
		for _, key := range keys {
			items = append(items, &DeleteStateItem{
				Key:      key,
				Metadata: map[string]string{},
				Options: &StateOptions{
					Concurrency: StateConcurrencyFirstWrite,
					Consistency: StateConsistencyEventual,
				},
			})
		}
		err := testClient.DeleteBulkStateItems(ctx, store, items)
		assert.Nil(t, err)
	})

	t.Run("delete exist data", func(t *testing.T) {
		// save data
		items := make([]*SetStateItem, 0, len(keys))
		for _, key := range keys {
			items = append(items, &SetStateItem{
				Key:      key,
				Value:    []byte(data),
				Metadata: map[string]string{},
				Etag:     &ETag{Value: "1"},
				Options: &StateOptions{
					Concurrency: StateConcurrencyFirstWrite,
					Consistency: StateConsistencyEventual,
				},
			})
		}
		err := testClient.SaveBulkState(ctx, store, items...)
		assert.Nil(t, err)

		// confirm data saved
		getItems, err := testClient.GetBulkState(ctx, store, keys, nil, 1)
		assert.NoError(t, err)
		assert.Equal(t, len(keys), len(getItems))

		// delete
		err = testClient.DeleteBulkState(ctx, store, keys, nil)
		assert.NoError(t, err)

		// confirm data deleted
		getItems, err = testClient.GetBulkState(ctx, store, keys, nil, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(getItems))
	})

	t.Run("delete exist data with stateItem", func(t *testing.T) {
		// save data
		items := make([]*SetStateItem, 0, len(keys))
		for _, key := range keys {
			items = append(items, &SetStateItem{
				Key:      key,
				Value:    []byte(data),
				Metadata: map[string]string{},
				Etag:     &ETag{Value: "1"},
				Options: &StateOptions{
					Concurrency: StateConcurrencyFirstWrite,
					Consistency: StateConsistencyEventual,
				},
			})
		}
		err := testClient.SaveBulkState(ctx, store, items...)
		assert.Nil(t, err)

		// confirm data saved
		getItems, err := testClient.GetBulkState(ctx, store, keys, nil, 1)
		assert.NoError(t, err)
		assert.Equal(t, len(keys), len(getItems))

		// delete
		deleteItems := make([]*DeleteStateItem, 0, len(keys))
		for _, key := range keys {
			deleteItems = append(deleteItems, &DeleteStateItem{
				Key:      key,
				Metadata: map[string]string{},
				Etag:     &ETag{Value: "1"},
				Options: &StateOptions{
					Concurrency: StateConcurrencyFirstWrite,
					Consistency: StateConsistencyEventual,
				},
			})
		}
		err = testClient.DeleteBulkStateItems(ctx, store, deleteItems)
		assert.Nil(t, err)

		// confirm data deleted
		getItems, err = testClient.GetBulkState(ctx, store, keys, nil, 1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(getItems))
	})
}

// go test -timeout 30s ./client -count 1 -run ^TestStateTransactions$
func TestStateTransactions(t *testing.T) {
	ctx := context.Background()
	data := `{ "message": "test" }`
	store := testStore
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
		items, err := testClient.GetBulkState(ctx, store, keys, nil, 10)
		assert.Nil(t, err)
		assert.NotNil(t, items)
		assert.Len(t, items, len(keys))

		upsers := make([]*StateOperation, 0)
		for _, item := range items {
			op := &StateOperation{
				Type: StateOperationTypeUpsert,
				Item: &SetStateItem{
					Key: item.Key,
					Etag: &ETag{
						Value: item.Etag,
					},
					Value: item.Value,
				},
			}
			upsers = append(upsers, op)
		}
		err = testClient.ExecuteStateTransaction(ctx, store, meta, upsers)
		assert.Nil(t, err)
	})

	t.Run("get and validate inserts", func(t *testing.T) {
		items, err := testClient.GetBulkState(ctx, store, keys, nil, 10)
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
		items, err := testClient.GetBulkState(ctx, store, keys, nil, 3)
		assert.Nil(t, err)
		assert.NotNil(t, items)
		assert.Len(t, items, 0)
	})
}

func TestQueryState(t *testing.T) {
	ctx := context.Background()
	data := testData
	store := testStore
	key1 := "key1"
	key2 := "key2"

	t.Run("save data", func(t *testing.T) {
		err := testClient.SaveState(ctx, store, key1, []byte(data), nil)
		assert.NoError(t, err)
		err = testClient.SaveState(ctx, store, key2, []byte(data), nil)
		assert.NoError(t, err)
	})

	t.Run("error query", func(t *testing.T) {
		_, err := testClient.QueryStateAlpha1(ctx, "", "", nil)
		assert.Error(t, err)
		_, err = testClient.QueryStateAlpha1(ctx, store, "", nil)
		assert.Error(t, err)
		_, err = testClient.QueryStateAlpha1(ctx, store, "bad syntax", nil)
		assert.Error(t, err)
	})

	t.Run("query data", func(t *testing.T) {
		query := `{}`
		resp, err := testClient.QueryStateAlpha1(ctx, store, query, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.Results))
		for _, item := range resp.Results {
			assert.True(t, item.Key == key1 || item.Key == key2)
			assert.Equal(t, []byte(data), item.Value)
		}
	})
}
