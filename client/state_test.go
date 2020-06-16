package client

import (
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
