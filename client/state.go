package client

import (
	"context"
	"time"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	duration "github.com/golang/protobuf/ptypes/duration"
	"github.com/pkg/errors"
)

const (
	// StateConsistencyUndefined is the undefined value for state consistency.
	StateConsistencyUndefined StateConsistency = 0
	// StateConsistencyEventual represents eventual state consistency value.
	StateConsistencyEventual StateConsistency = 1
	// StateConsistencyStrong represents strong state consistency value.
	StateConsistencyStrong StateConsistency = 2

	// StateConcurrencyUndefined is the undefined value for state concurrency.
	StateConcurrencyUndefined StateConcurrency = 0
	// StateConcurrencyFirstWrite represents first write concurrency value.
	StateConcurrencyFirstWrite StateConcurrency = 1
	// StateConcurrencyLastWrite represents last write concurrency value.
	StateConcurrencyLastWrite StateConcurrency = 2

	// RetryPatternUndefined is the undefined value for retry pattern.
	RetryPatternUndefined RetryPattern = 0
	// RetryPatternLinear represents the linear retry pattern value.
	RetryPatternLinear RetryPattern = 1
	// RetryPatternExponential represents the exponential retry pattern value.
	RetryPatternExponential RetryPattern = 2
)

type (
	// StateConsistency is the consistency enum type.
	StateConsistency int
	// StateConcurrency is the concurrency enum type.
	StateConcurrency int
	// RetryPattern is the retry pattern enum type.
	RetryPattern int
)

// String returns the string value of the StateConsistency.
func (c StateConsistency) String() string {
	names := [...]string{
		"Undefined",
		"Strong",
		"Eventual",
	}
	if c < StateConsistencyStrong || c > StateConsistencyEventual {
		return "Undefined"
	}

	return names[c]
}

// String returns the string value of the StateConcurrency.
func (c StateConcurrency) String() string {
	names := [...]string{
		"Undefined",
		"FirstWrite",
		"LastWrite",
	}
	if c < StateConcurrencyFirstWrite || c > StateConcurrencyLastWrite {
		return "Undefined"
	}

	return names[c]
}

// String returns the string value of the RetryPattern.
func (c RetryPattern) String() string {
	names := [...]string{
		"Undefined",
		"Linear",
		"Exponential",
	}
	if c < RetryPatternLinear || c > RetryPatternExponential {
		return "Undefined"
	}

	return names[c]
}

var (
	stateOptionRetryPolicyDefault = &v1.StateRetryPolicy{
		Threshold: 3,
		Pattern:   v1.StateRetryPolicy_RETRY_EXPONENTIAL,
	}

	stateOptionDefault = &v1.StateOptions{
		Concurrency: v1.StateOptions_CONCURRENCY_LAST_WRITE,
		Consistency: v1.StateOptions_CONSISTENCY_STRONG,
		RetryPolicy: stateOptionRetryPolicyDefault,
	}
)

// State is a collection of StateItems with a store name.
type State struct {
	StoreName string
	States    []*StateItem
}

// StateItem represents a single state to be persisted.
type StateItem struct {
	Key      string
	Value    []byte
	Etag     string
	Metadata map[string]string
	Options  *StateOptions
}

// StateOptions represents the state store persistence policy.
type StateOptions struct {
	Concurrency StateConcurrency
	Consistency StateConsistency
	RetryPolicy *StateRetryPolicy
}

// StateRetryPolicy represents the state store invocation retry policy.
type StateRetryPolicy struct {
	Threshold int32
	Pattern   RetryPattern
	Interval  time.Duration
}

func toProtoSaveStateRequest(s *State) (req *pb.SaveStateRequest) {
	r := &pb.SaveStateRequest{
		StoreName: s.StoreName,
		States:    make([]*v1.StateItem, 0),
	}

	for _, si := range s.States {
		item := toProtoSaveStateItem(si)
		r.States = append(r.States, item)
	}
	return r
}

func toProtoSaveStateItem(si *StateItem) (item *v1.StateItem) {
	return &v1.StateItem{
		Etag:     si.Etag,
		Key:      si.Key,
		Metadata: si.Metadata,
		Value:    si.Value,
		Options:  toProtoStateOptions(si.Options),
	}
}

func toProtoStateOptions(so *StateOptions) (opts *v1.StateOptions) {
	if so == nil {
		return stateOptionDefault
	}
	return &v1.StateOptions{
		Concurrency: (v1.StateOptions_StateConcurrency(so.Concurrency)),
		Consistency: (v1.StateOptions_StateConsistency(so.Consistency)),
		RetryPolicy: &v1.StateRetryPolicy{
			Interval:  toProtoDuration(so.RetryPolicy.Interval),
			Pattern:   (v1.StateRetryPolicy_RetryPattern(so.RetryPolicy.Pattern)),
			Threshold: so.RetryPolicy.Threshold,
		},
	}
}

func toProtoDuration(d time.Duration) *duration.Duration {
	nanos := d.Nanoseconds()
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return &duration.Duration{
		Seconds: int64(secs),
		Nanos:   int32(nanos),
	}
}

// SaveState saves the fully loaded state to store.
func (c *GRPCClient) SaveState(ctx context.Context, s *State) error {
	if s == nil || s.StoreName == "" || s.States == nil || len(s.States) < 1 {
		return errors.New("nil or invalid state")
	}
	req := toProtoSaveStateRequest(s)
	_, err := c.protoClient.SaveState(authContext(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error saving state")
	}
	return nil
}

// SaveStateDataVersion saves the raw data into store using default state options and etag.
func (c *GRPCClient) SaveStateDataVersion(ctx context.Context, store, key, etag string, data []byte) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}

	req := &State{
		StoreName: store,
		States: []*StateItem{
			{
				Key:   key,
				Value: data,
				Etag:  etag,
			},
		},
	}

	return c.SaveState(ctx, req)
}

// SaveStateData saves the raw data into store using default state options.
func (c *GRPCClient) SaveStateData(ctx context.Context, store, key string, data []byte) error {
	return c.SaveStateDataVersion(ctx, store, key, "", data)
}

// SaveStateItem saves the single state item to store.
func (c *GRPCClient) SaveStateItem(ctx context.Context, store string, item *StateItem) error {
	if store == "" {
		return errors.New("nil store")
	}
	if item == nil {
		return errors.New("nil item")
	}

	req := &State{
		StoreName: store,
		States:    []*StateItem{item},
	}

	return c.SaveState(ctx, req)
}

// GetState retreaves state from specific store using default consistency option.
func (c *GRPCClient) GetState(ctx context.Context, store, key string) (out []byte, etag string, err error) {
	return c.GetStateWithConsistency(ctx, store, key, StateConsistencyStrong)
}

// GetStateWithConsistency retreaves state from specific store using provided state consistency.
func (c *GRPCClient) GetStateWithConsistency(ctx context.Context, store, key string, sc StateConsistency) (out []byte, etag string, err error) {
	if store == "" {
		return nil, "", errors.New("nil store")
	}
	if key == "" {
		return nil, "", errors.New("nil key")
	}

	req := &pb.GetStateRequest{
		StoreName:   store,
		Key:         key,
		Consistency: (v1.StateOptions_StateConsistency(sc)),
	}

	result, err := c.protoClient.GetState(authContext(ctx), req)
	if err != nil {
		return nil, "", errors.Wrap(err, "error getting state")
	}

	return result.Data, result.Etag, nil
}

// DeleteState deletes content from store using default state options.
func (c *GRPCClient) DeleteState(ctx context.Context, store, key string) error {
	return c.DeleteStateVersion(ctx, store, key, "", nil)
}

// DeleteStateVersion deletes content from store using provided state options and etag.
func (c *GRPCClient) DeleteStateVersion(ctx context.Context, store, key, etag string, opts *StateOptions) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}

	req := &pb.DeleteStateRequest{
		StoreName: store,
		Key:       key,
		Etag:      etag,
		Options:   toProtoStateOptions(opts),
	}

	_, err := c.protoClient.DeleteState(authContext(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error deleting state")
	}

	return nil
}
