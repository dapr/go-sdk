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

	// StateOperationTypeUndefined is the undefined value for state operation type.
	StateOperationTypeUndefined OperationType = 0
	// StateOperationTypeUpsert represents upsert operation type value.
	StateOperationTypeUpsert OperationType = 1
	// StateOperationTypeDelete represents delete operation type value.
	StateOperationTypeDelete OperationType = 2
)

type (
	// StateConsistency is the consistency enum type.
	StateConsistency int
	// StateConcurrency is the concurrency enum type.
	StateConcurrency int
	// OperationType is the operation enum type.
	OperationType int
)

// String returns the string value of the OperationType.
func (o OperationType) String() string {
	names := [...]string{
		"undefined",
		"upsert",
		"delete",
	}
	if o < StateOperationTypeUpsert || o > StateOperationTypeDelete {
		return "undefined"
	}

	return names[o]
}

// String returns the string value of the StateConsistency.
func (c StateConsistency) String() string {
	names := [...]string{
		"undefined",
		"strong",
		"eventual",
	}
	if c < StateConsistencyStrong || c > StateConsistencyEventual {
		return "undefined"
	}

	return names[c]
}

// String returns the string value of the StateConcurrency.
func (c StateConcurrency) String() string {
	names := [...]string{
		"undefined",
		"first-write",
		"last-write",
	}
	if c < StateConcurrencyFirstWrite || c > StateConcurrencyLastWrite {
		return "undefined"
	}

	return names[c]
}

var (
	stateOptionDefault = &v1.StateOptions{
		Concurrency: v1.StateOptions_CONCURRENCY_LAST_WRITE,
		Consistency: v1.StateOptions_CONSISTENCY_STRONG,
	}
)

// StateOperation is a collection of StateItems with a store name.
type StateOperation struct {
	Type OperationType
	Item *SetStateItem
}

// StateItem represents a single state item.
type StateItem struct {
	Key   string
	Value []byte
	Etag  string
}

// SetStateItem represents a single state to be persisted.
type SetStateItem struct {
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
}

func toProtoSaveStateItem(si *SetStateItem) (item *v1.StateItem) {
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

// ExecuteStateTransaction provides way to execute multiple operations on a specified store.
func (c *GRPCClient) ExecuteStateTransaction(ctx context.Context, store string, meta map[string]string, ops []*StateOperation) error {
	if store == "" {
		return errors.New("nil store")
	}
	if len(ops) == 0 {
		return nil
	}

	items := make([]*pb.TransactionalStateOperation, 0)
	for _, op := range ops {
		item := &pb.TransactionalStateOperation{
			OperationType: op.Type.String(),
			Request:       toProtoSaveStateItem(op.Item),
		}
		items = append(items, item)
	}

	req := &pb.ExecuteStateTransactionRequest{
		Metadata:   meta,
		StoreName:  store,
		Operations: items,
	}
	_, err := c.protoClient.ExecuteStateTransaction(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error executing state transaction")
	}
	return nil
}

// SaveState saves the raw data into store using default state options.
func (c *GRPCClient) SaveState(ctx context.Context, store, key string, data []byte) error {
	item := &SetStateItem{Key: key, Value: data}
	return c.SaveStateItems(ctx, store, item)
}

// SaveStateItems saves the multiple state item to store.
func (c *GRPCClient) SaveStateItems(ctx context.Context, store string, items ...*SetStateItem) error {
	if store == "" {
		return errors.New("nil store")
	}
	if items == nil {
		return errors.New("nil item")
	}

	req := &pb.SaveStateRequest{
		StoreName: store,
		States:    make([]*v1.StateItem, 0),
	}

	for _, si := range items {
		item := toProtoSaveStateItem(si)
		req.States = append(req.States, item)
	}

	_, err := c.protoClient.SaveState(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error saving state")
	}
	return nil

}

// GetBulkItems retreaves state for multiple keys from specific store.
func (c *GRPCClient) GetBulkItems(ctx context.Context, store string, keys []string, parallelism int32) ([]*StateItem, error) {
	if store == "" {
		return nil, errors.New("nil store")
	}
	if len(keys) == 0 {
		return nil, errors.New("keys required")
	}
	items := make([]*StateItem, 0)

	req := &pb.GetBulkStateRequest{
		StoreName:   store,
		Keys:        keys,
		Parallelism: parallelism,
	}

	results, err := c.protoClient.GetBulkState(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting state")
	}

	if results == nil || results.Items == nil {
		return items, nil
	}

	for _, r := range results.Items {
		item := &StateItem{
			Key:   r.Key,
			Etag:  r.Etag,
			Value: r.Data,
		}
		items = append(items, item)
	}

	return items, nil
}

// GetState retreaves state from specific store using default consistency option.
func (c *GRPCClient) GetState(ctx context.Context, store, key string) (item *StateItem, err error) {
	return c.GetStateWithConsistency(ctx, store, key, nil, StateConsistencyStrong)
}

// GetStateWithConsistency retreaves state from specific store using provided state consistency.
func (c *GRPCClient) GetStateWithConsistency(ctx context.Context, store, key string, meta map[string]string, sc StateConsistency) (item *StateItem, err error) {
	if store == "" {
		return nil, errors.New("nil store")
	}
	if key == "" {
		return nil, errors.New("nil key")
	}

	req := &pb.GetStateRequest{
		StoreName:   store,
		Key:         key,
		Consistency: (v1.StateOptions_StateConsistency(sc)),
		Metadata:    meta,
	}

	result, err := c.protoClient.GetState(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting state")
	}

	return &StateItem{
		Etag:  result.Etag,
		Key:   key,
		Value: result.Data,
	}, nil
}

// DeleteState deletes content from store using default state options.
func (c *GRPCClient) DeleteState(ctx context.Context, store, key string) error {
	return c.DeleteStateWithETag(ctx, store, key, "", nil, nil)
}

// DeleteStateWithETag deletes content from store using provided state options and etag.
func (c *GRPCClient) DeleteStateWithETag(ctx context.Context, store, key, etag string, meta map[string]string, opts *StateOptions) error {
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
		Metadata:  meta,
	}

	_, err := c.protoClient.DeleteState(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error deleting state")
	}

	return nil
}
