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
		RetryPolicy: &StateRetryPolicy{
			Threshold: 3,
			Interval:  time.Duration(10 * time.Second),
			Pattern:   RetryPatternExponential,
		},
	}
	p := toProtoStateOptions(s)
	assert.NotNil(t, p)
	assert.Equal(t, p.Concurrency, v1.StateOptions_CONCURRENCY_LAST_WRITE)
	assert.Equal(t, p.Consistency, v1.StateOptions_CONSISTENCY_STRONG)
	assert.NotNil(t, p.RetryPolicy)
	assert.Equal(t, p.RetryPolicy.Threshold, int32(3))
	assert.Equal(t, p.RetryPolicy.Interval.Seconds, int64(10))
	assert.Equal(t, p.RetryPolicy.Pattern, v1.StateRetryPolicy_RETRY_EXPONENTIAL)
}

func TestSaveStateData(t *testing.T) {
	ctx := context.Background()
	data := "test"

	err := testClient.SaveStateData(ctx, "store", "key1", []byte(data))
	assert.Nil(t, err)

	out, etag, err := testClient.GetState(ctx, "store", "key1")
	assert.Nil(t, err)
	assert.NotEmpty(t, etag)
	assert.NotNil(t, out)
	assert.Equal(t, string(out), data)

	err = testClient.SaveStateDataVersion(ctx, "store", "key1", etag, []byte(data))
	assert.Nil(t, err)

	err = testClient.DeleteState(ctx, "store", "key1")
	assert.Nil(t, err)
}
