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
		err := testClient.SaveStateData(ctx, store, key, []byte(data))
		assert.Nil(t, err)
	})

	t.Run("get saved data", func(t *testing.T) {
		out, etag, err := testClient.GetState(ctx, store, key)
		assert.Nil(t, err)
		assert.NotEmpty(t, etag)
		assert.NotNil(t, out)
		assert.Equal(t, string(out), data)
	})

	t.Run("save data with version", func(t *testing.T) {
		err := testClient.SaveStateDataWithETag(ctx, store, key, "1", []byte(data))
		assert.Nil(t, err)
	})

	t.Run("delete data", func(t *testing.T) {
		err := testClient.DeleteState(ctx, store, key)
		assert.Nil(t, err)
	})
}
