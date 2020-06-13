package client

import (
	"context"
	"encoding/json"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

var (
	// StateOptionConsistencyDefault is strong
	StateOptionConsistencyDefault = v1.StateOptions_CONSISTENCY_STRONG

	// StateOptionConcurrencyDefault is last write
	StateOptionConcurrencyDefault = v1.StateOptions_CONCURRENCY_LAST_WRITE

	// StateOptionRetryPolicyDefault is threshold 3
	StateOptionRetryPolicyDefault = &v1.StateRetryPolicy{
		Threshold: 3,
	}

	// StateOptionDefault is the optimistic state option (last write concurency and strong consistency)
	StateOptionDefault = &v1.StateOptions{
		Concurrency: StateOptionConcurrencyDefault,
		Consistency: StateOptionConsistencyDefault,
		RetryPolicy: StateOptionRetryPolicyDefault,
	}
)

// *** Save State ***

// SaveState saves the fully loaded save state request
func (c *Client) SaveState(ctx context.Context, req *pb.SaveStateRequest) error {
	if req == nil {
		return errors.New("nil request")
	}

	_, err := c.protoClient.SaveState(authContext(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error saving state")
	}

	return nil
}

// SaveStateItem saves a single state item
func (c *Client) SaveStateItem(ctx context.Context, store string, item *v1.StateItem) error {
	if store == "" {
		return errors.New("nil store")
	}
	if item == nil {
		return errors.New("nil item")
	}

	req := &pb.SaveStateRequest{
		StoreName: store,
		States:    []*v1.StateItem{item},
	}

	return c.SaveState(ctx, req)
}

// SaveStateWithData saves the data into store using default state options
func (c *Client) SaveStateWithData(ctx context.Context, store, key string, data []byte) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}

	req := &pb.SaveStateRequest{
		StoreName: store,
		States: []*v1.StateItem{
			{
				Key:     key,
				Value:   data,
				Options: StateOptionDefault,
			},
		},
	}

	return c.SaveState(ctx, req)
}

// SaveStateJSON saves the JSON serialized in into store using default state options
func (c *Client) SaveStateJSON(ctx context.Context, store, key string, in interface{}) error {
	if in == nil {
		return errors.New("nil data to save")
	}
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "error marshaling content")
	}
	return c.SaveStateWithData(ctx, store, key, b)
}

// *** Get State ***

// GetStateWithRequest retreaves state from specific store using provided request
func (c *Client) GetStateWithRequest(ctx context.Context, req *pb.GetStateRequest) (out []byte, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	result, err := c.protoClient.GetState(authContext(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting state")
	}

	return result.Data, nil
}

// GetState retreaves state from specific store using default consistency option
func (c *Client) GetState(ctx context.Context, store, key string) (out []byte, err error) {
	if store == "" {
		return nil, errors.New("nil store")
	}
	if key == "" {
		return nil, errors.New("nil key")
	}
	req := &pb.GetStateRequest{
		StoreName:   store,
		Key:         key,
		Consistency: StateOptionConsistencyDefault,
	}

	return c.GetStateWithRequest(ctx, req)
}

// *** Delete State ***

// DeleteStateWithRequest deletes content from store using provided request
func (c *Client) DeleteStateWithRequest(ctx context.Context, req *pb.DeleteStateRequest) error {
	if req == nil {
		return errors.New("nil request")
	}

	_, err := c.protoClient.DeleteState(authContext(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error deleting state")
	}

	return nil
}

// DeleteState deletes content from store using default state options
func (c *Client) DeleteState(ctx context.Context, store, key string) error {
	if store == "" {
		return errors.New("nil store")
	}
	if key == "" {
		return errors.New("nil key")
	}
	req := &pb.DeleteStateRequest{
		StoreName: store,
		Key:       key,
		Options:   StateOptionDefault,
	}

	return c.DeleteStateWithRequest(ctx, req)
}
