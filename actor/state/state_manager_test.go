/*
Copyright 2026 The Dapr Authors
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

package state

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/go-sdk/client"
)

// fakeActorStateClient serves actor state from an in-memory map and records
// the transactional operations it receives.
type fakeActorStateClient struct {
	client.Client
	data map[string][]byte
	err  error
	ops  []*client.ActorStateOperation
}

func (f *fakeActorStateClient) GetActorState(_ context.Context, in *client.GetActorStateRequest) (*client.GetActorStateResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &client.GetActorStateResponse{Data: f.data[in.KeyName]}, nil
}

func (f *fakeActorStateClient) SaveStateTransactionally(_ context.Context, _, _ string, operations []*client.ActorStateOperation) error {
	f.ops = append(f.ops, operations...)
	return nil
}

func TestStateManagerRemoveUncachedState(t *testing.T) {
	ctx := context.Background()
	fake := &fakeActorStateClient{data: map[string][]byte{"key": []byte(`"value"`)}}
	sm := NewActorStateManagerContext("testActor", "test-0", NewDaprStateAsyncProvider(fake))

	require.NoError(t, sm.Remove(ctx, "key"))

	ok, err := sm.Contains(ctx, "key")
	require.NoError(t, err)
	assert.False(t, ok)

	require.NoError(t, sm.Save(ctx))
	require.Len(t, fake.ops, 1)
	assert.Equal(t, string(Remove), fake.ops[0].OperationType)
	assert.Equal(t, "key", fake.ops[0].Key)
}

func TestStateManagerRemoveMissingState(t *testing.T) {
	ctx := context.Background()
	fake := &fakeActorStateClient{data: map[string][]byte{}}
	sm := NewActorStateManagerContext("testActor", "test-0", NewDaprStateAsyncProvider(fake))

	require.NoError(t, sm.Remove(ctx, "key"))
	require.NoError(t, sm.Save(ctx))
	assert.Empty(t, fake.ops)
}

func TestStateManagerRemoveLookupError(t *testing.T) {
	ctx := context.Background()
	fake := &fakeActorStateClient{err: errors.New("unavailable")}
	sm := NewActorStateManagerContext("testActor", "test-0", NewDaprStateAsyncProvider(fake))

	require.Error(t, sm.Remove(ctx, "key"))
}
