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

package state

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dapr/go-sdk/actor/mock"
	"github.com/dapr/go-sdk/actor/mock_client"
	"github.com/dapr/go-sdk/client"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testState = "state"
	testValue = "value"
	testTTL   = time.Second * 3600
)

func newMockStateManager(t *testing.T) *stateManager {
	return &stateManager{
		stateManagerCtx: newMockStateManagerCtx(t),
	}
}

func newMockStateManagerCtx(t *testing.T) *stateManagerCtx {
	ctrl := gomock.NewController(t)
	return &stateManagerCtx{
		actorTypeName: "test",
		actorID:       "fn",
		stateAsyncProvider: &DaprStateAsyncProvider{
			daprClient:      mock_client.NewMockClient(ctrl),
			stateSerializer: mock.NewMockCodec(ctrl),
		},
	}
}

func newGetActorStateRequest(sm *stateManagerCtx, key string) *client.GetActorStateRequest {
	return &client.GetActorStateRequest{
		ActorType: sm.actorTypeName,
		ActorID:   sm.actorID,
		KeyName:   key,
	}
}

func newGetActorStateResponse(data []byte) *client.GetActorStateResponse {
	return &client.GetActorStateResponse{Data: data}
}

func TestStateManagerWithContext(t *testing.T) {
	sm := newMockStateManager(t)
	assert.Equal(t, sm.WithContext(), sm.stateManagerCtx)
}

func TestAdd_EmptyStateName(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManager(t)
	err := sm.Add("", testValue)
	require.Error(t, err)
	err = sm.stateManagerCtx.Add(ctx, "", testValue)
	require.Error(t, err)
}

func TestAdd_WithCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		kind      ChangeKind
		shouldErr bool
		context   context.Context
	}{
		{"state change kind None", None, true, nil},
		{"state change kind Add", Add, true, nil},
		{"state change kind Update", Update, true, nil},
		{"state change kind Remove", Remove, false, nil},
		{"context state change kind None", None, true, ctx},
		{"context state change kind Add", Add, true, ctx},
		{"context state change kind Update", Update, true, ctx},
		{"context state change kind Remove", Remove, false, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			sm.stateChangeTracker.Store(testState, &ChangeMetadata{Kind: tt.kind, Value: testValue})
			mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
			mockRequest := newGetActorStateRequest(sm.stateManagerCtx, testState)
			mockResult := newGetActorStateResponse([]byte("result"))
			mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(mockResult, nil)

			var err error
			if tt.context == nil {
				err = sm.Add(testState, testValue)
			} else {
				err = sm.stateManagerCtx.Add(tt.context, testState, testValue)
			}

			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				val, ok := sm.stateChangeTracker.Load(testState)
				require.True(t, ok)

				metadata := val.(*ChangeMetadata)
				assert.Equal(t, Update, metadata.Kind)
				assert.Equal(t, testValue, metadata.Value)
			}
		})
	}
}

func TestAdd_WithoutCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		stateProviderErr bool
		duplicateState   bool
		context          context.Context
	}{
		{"state provider returns error", true, false, nil},
		{"state provider returns data", false, true, nil},
		{"successfully add new state", false, false, nil},
		{"context state provider returns error", true, false, ctx},
		{"context state provider returns data", false, true, ctx},
		{"context successfully add new state", false, false, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
			mockRequest := newGetActorStateRequest(sm.stateManagerCtx, testState)
			mockResult := newGetActorStateResponse([]byte("result"))
			if tt.stateProviderErr {
				mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(nil, errors.New("mockErr"))
			} else {
				if tt.duplicateState {
					mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(mockResult, nil)
				} else {
					mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(nil, nil)
				}
			}

			var err error
			if tt.context == nil {
				err = sm.Add(testState, testValue)
			} else {
				err = sm.stateManagerCtx.Add(tt.context, testState, testValue)
			}

			if tt.stateProviderErr || tt.duplicateState {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				val, ok := sm.stateChangeTracker.Load(testState)
				require.True(t, ok)

				metadata := val.(*ChangeMetadata)
				assert.Equal(t, Add, metadata.Kind)
				assert.Equal(t, testValue, metadata.Value)
			}
		})
	}
}

func TestGet_EmptyStateName(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManager(t)
	err := sm.Get("", testValue)
	require.Error(t, err)
	err = sm.stateManagerCtx.Get(ctx, "", testValue)
	require.Error(t, err)
}

func TestGet_WithCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		kind      ChangeKind
		shouldErr bool
		context   context.Context
	}{
		{"state change kind None", None, false, nil},
		{"state change kind Add", Add, false, nil},
		{"state change kind Update", Update, false, nil},
		{"state change kind Remove", Remove, true, nil},
		{"context state change kind None", None, false, ctx},
		{"context state change kind Add", Add, false, ctx},
		{"context state change kind Update", Update, false, ctx},
		{"context state change kind Remove", Remove, true, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			sm.stateChangeTracker.Store(testState, &ChangeMetadata{Kind: tt.kind, Value: testValue})

			var err error
			var reply string
			if tt.context == nil {
				err = sm.Get(testState, &reply)
			} else {
				err = sm.stateManagerCtx.Get(tt.context, testState, &reply)
			}

			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testValue, reply)
			}
		})
	}
}

func TestGet_WithoutCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		shouldErr bool
		context   context.Context
	}{
		{"state provider returns error", true, nil},
		{"state provider returns data", false, nil},
		{"context state provider returns error", true, ctx},
		{"context state provider returns data", false, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
			mockCodec := sm.stateAsyncProvider.stateSerializer.(*mock.MockCodec)
			mockRequest := newGetActorStateRequest(sm.stateManagerCtx, testState)
			mockResult := newGetActorStateResponse([]byte("result"))
			if tt.shouldErr {
				mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(nil, errors.New("mockErr"))
			} else {
				mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(mockResult, nil)
				mockCodec.EXPECT().Unmarshal(mockResult.Data, testValue)
			}

			var err error
			if tt.context == nil {
				err = sm.Get(testState, testValue)
			} else {
				err = sm.stateManagerCtx.Get(tt.context, testState, testValue)
			}

			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			val, ok := sm.stateChangeTracker.Load(testState)
			require.True(t, ok)

			metadata := val.(*ChangeMetadata)
			assert.Equal(t, None, metadata.Kind)
			assert.Equal(t, testValue, metadata.Value)
		})
	}
}

func TestSet_EmptyStateName(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManager(t)
	err := sm.Set("", testValue)
	require.Error(t, err)
	err = sm.stateManagerCtx.Set(ctx, "", testValue)
	require.Error(t, err)
}

func TestSet_WithCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		initKind   ChangeKind
		expectKind ChangeKind
		context    context.Context
	}{
		{"state change kind None", None, Update, nil},
		{"state change kind Add", Add, Add, nil},
		{"state change kind Update", Update, Update, nil},
		{"state change kind Remove", Remove, Update, nil},
		{"context state change kind None", None, Update, ctx},
		{"context state change kind Add", Add, Add, ctx},
		{"context state change kind Update", Update, Update, ctx},
		{"context state change kind Remove", Remove, Update, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			sm.stateChangeTracker.Store(testState, &ChangeMetadata{Kind: tt.initKind, Value: testValue})

			if tt.context == nil {
				err := sm.Set(testState, testValue)
				require.NoError(t, err)
			} else {
				err := sm.stateManagerCtx.Set(tt.context, testState, testValue)
				require.NoError(t, err)
			}

			val, ok := sm.stateChangeTracker.Load(testState)
			require.True(t, ok)

			metadata := val.(*ChangeMetadata)
			assert.Equal(t, tt.expectKind, metadata.Kind)
			assert.Equal(t, testValue, metadata.Value)
		})
	}
}

func TestSet_WithoutCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		context context.Context
	}{
		{"without context", nil},
		{"with context", ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)

			if tt.context == nil {
				err := sm.Set(testState, testValue)
				require.NoError(t, err)
			} else {
				err := sm.stateManagerCtx.Set(tt.context, testState, testValue)
				require.NoError(t, err)
			}

			val, ok := sm.stateChangeTracker.Load(testState)
			require.True(t, ok)

			metadata := val.(*ChangeMetadata)
			assert.Equal(t, Add, metadata.Kind)
			assert.Equal(t, testValue, metadata.Value)
		})
	}
}

func TestSetWithTTL_EmptyStateName(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManagerCtx(t)
	err := sm.SetWithTTL(ctx, "", testValue, testTTL)
	require.Error(t, err)
}

func TestSetWithTTL_NegativeTTL(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManagerCtx(t)
	err := sm.SetWithTTL(ctx, testState, testValue, -testTTL)
	require.Error(t, err)
}

func TestSetWithTTL_WithCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		initKind   ChangeKind
		expectKind ChangeKind
	}{
		{"state change kind None", None, Update},
		{"state change kind Add", Add, Add},
		{"state change kind Update", Update, Update},
		{"state change kind Remove", Remove, Update},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManagerCtx(t)
			sm.stateChangeTracker.Store(testState, &ChangeMetadata{Kind: tt.initKind, Value: testValue})

			err := sm.SetWithTTL(ctx, testState, testValue, testTTL)
			require.NoError(t, err)

			val, ok := sm.stateChangeTracker.Load(testState)
			require.True(t, ok)

			metadata := val.(*ChangeMetadata)
			assert.Equal(t, tt.expectKind, metadata.Kind)
			assert.Equal(t, testValue, metadata.Value)
			assert.Equal(t, testTTL, *metadata.TTL)
		})
	}
}

func TestSetWithTTL_WithoutCachedStateChange(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManagerCtx(t)

	err := sm.SetWithTTL(ctx, testState, testValue, testTTL)
	require.NoError(t, err)

	val, ok := sm.stateChangeTracker.Load(testState)
	require.True(t, ok)

	metadata := val.(*ChangeMetadata)
	assert.Equal(t, Add, metadata.Kind)
	assert.Equal(t, testValue, metadata.Value)
	assert.Equal(t, testTTL, *metadata.TTL)
}

func TestRemove_EmptyStateName(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManager(t)
	err := sm.Remove("")
	require.Error(t, err)
	err = sm.stateManagerCtx.Remove(ctx, "")
	require.Error(t, err)
}

func TestRemove_WithCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		kind    ChangeKind
		inCache bool
		context context.Context
	}{
		{"state change kind None", None, true, nil},
		{"state change kind Add", Add, false, nil},
		{"state change kind Update", Update, false, nil},
		{"state change kind Remove", Remove, true, nil},
		{"context state change kind None", None, true, ctx},
		{"context state change kind Add", Add, false, ctx},
		{"context state change kind Update", Update, false, ctx},
		{"context state change kind Remove", Remove, true, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			sm.stateChangeTracker.Store(testState, &ChangeMetadata{Kind: tt.kind, Value: testValue})

			if tt.context == nil {
				err := sm.Remove(testState)
				require.NoError(t, err)
			} else {
				err := sm.stateManagerCtx.Remove(tt.context, testState)
				require.NoError(t, err)
			}

			val, ok := sm.stateChangeTracker.Load(testState)
			if tt.inCache {
				assert.Equal(t, Remove, val.(*ChangeMetadata).Kind)
				assert.True(t, ok)
			} else {
				assert.Nil(t, val)
				assert.False(t, ok)
			}
		})
	}
}

func TestRemove_WithoutCachedStateChange(t *testing.T) {
	ctx := context.Background()
	mockResult := newGetActorStateResponse([]byte("result"))
	mockErr := errors.New("mockErr")

	tests := []struct {
		name     string
		mockFunc func(sm *stateManagerCtx, mc *mock_client.MockClient)
		context  context.Context
	}{
		{
			"state provider returns error",
			func(sm *stateManagerCtx, mc *mock_client.MockClient) {
				mc.EXPECT().GetActorState(ctx, newGetActorStateRequest(sm, testState)).Return(nil, mockErr)
			},
			nil,
		},
		{
			"state provider returns data",
			func(sm *stateManagerCtx, mc *mock_client.MockClient) {
				mc.EXPECT().GetActorState(ctx, newGetActorStateRequest(sm, testState)).Return(mockResult, nil)
			},
			nil},
		{
			"state provider returns error",
			func(sm *stateManagerCtx, mc *mock_client.MockClient) {
				mc.EXPECT().GetActorState(ctx, newGetActorStateRequest(sm, testState)).Return(nil, mockErr)
			},
			ctx},
		{
			"state provider returns data",
			func(sm *stateManagerCtx, mc *mock_client.MockClient) {
				mc.EXPECT().GetActorState(ctx, newGetActorStateRequest(sm, testState)).Return(mockResult, nil)
			},
			ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
			tt.mockFunc(sm.stateManagerCtx, mockClient)

			if tt.context == nil {
				err := sm.Remove(testState)
				require.NoError(t, err)
			} else {
				err := sm.stateManagerCtx.Remove(tt.context, testState)
				require.NoError(t, err)
			}
		})
	}
}

func TestContains_EmptyStateName(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManager(t)
	res, err := sm.Contains("")
	require.Error(t, err)
	assert.False(t, res)
	res, err = sm.stateManagerCtx.Contains(ctx, "")
	require.Error(t, err)
	assert.False(t, res)
}

func TestContains_WithCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		kind     ChangeKind
		expected bool
		context  context.Context
	}{
		{"state change kind None", None, true, nil},
		{"state change kind Add", Add, true, nil},
		{"state change kind Update", Update, true, nil},
		{"state change kind Remove", Remove, false, nil},
		{"context state change kind None", None, true, ctx},
		{"context state change kind Add", Add, true, ctx},
		{"context state change kind Update", Update, true, ctx},
		{"context state change kind Remove", Remove, false, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			sm.stateChangeTracker.Store(testState, &ChangeMetadata{Kind: tt.kind, Value: testValue})

			if tt.context == nil {
				result, err := sm.Contains(testState)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			} else {
				result, err := sm.stateManagerCtx.Contains(tt.context, testState)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestContains_WithoutCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		shouldErr bool
		context   context.Context
	}{
		{"state provider returns error", true, nil},
		{"state provider returns data", false, nil},
		{"context state provider returns error", true, ctx},
		{"context state provider returns data", false, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
			mockRequest := newGetActorStateRequest(sm.stateManagerCtx, testState)
			mockResult := newGetActorStateResponse([]byte("result"))
			if tt.shouldErr {
				mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(nil, errors.New("mockErr"))
			} else {
				mockClient.EXPECT().GetActorState(ctx, mockRequest).Return(mockResult, nil)
			}

			var result bool
			var err error
			if tt.context == nil {
				result, err = sm.Contains(testState)
			} else {
				result, err = sm.stateManagerCtx.Contains(tt.context, testState)
			}

			if tt.shouldErr {
				require.Error(t, err)
				assert.False(t, result)
			} else {
				require.NoError(t, err)
				assert.True(t, result)
			}
		})
	}
}

func TestSave_SingleCachedStateChange(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		stateChanges *ChangeMetadata
		context      context.Context
	}{
		{"no state change", nil, nil},
		{"state change kind None", &ChangeMetadata{Kind: None, Value: testValue}, nil},
		{"state change kind Add", &ChangeMetadata{Kind: Add, Value: testValue}, nil},
		{"state change kind Update", &ChangeMetadata{Kind: Update, Value: testValue}, nil},
		{"state change kind Remove", &ChangeMetadata{Kind: Remove, Value: testValue}, nil},
		{"context state change kind None", &ChangeMetadata{Kind: None, Value: testValue}, ctx},
		{"context state change kind Add", &ChangeMetadata{Kind: Add, Value: testValue}, ctx},
		{"context state change kind Update", &ChangeMetadata{Kind: Update, Value: testValue}, ctx},
		{"context state change kind Remove", &ChangeMetadata{Kind: Remove, Value: testValue}, ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
			mockCodec := sm.stateAsyncProvider.stateSerializer.(*mock.MockCodec)
			if tt.stateChanges != nil {
				sm.stateChangeTracker.Store(testState, tt.stateChanges)

				if tt.stateChanges.Kind == Remove {
					mockClient.EXPECT().SaveStateTransactionally(ctx, sm.actorTypeName, sm.actorID, gomock.Len(1))
				} else if tt.stateChanges.Kind == Add || tt.stateChanges.Kind == Update {
					mockClient.EXPECT().SaveStateTransactionally(ctx, sm.actorTypeName, sm.actorID, gomock.Len(1))
					mockCodec.EXPECT().Marshal(tt.stateChanges.Value)
				}
			}

			if tt.context == nil {
				err := sm.Save()
				require.NoError(t, err)
			} else {
				err := sm.stateManagerCtx.Save(tt.context)
				require.NoError(t, err)
			}
		})
	}
}

func TestSave_MultipleCachedStateChanges(t *testing.T) {
	ctx := context.Background()
	sm := newMockStateManager(t)
	mockClient := sm.stateAsyncProvider.daprClient.(*mock_client.MockClient)
	mockCodec := sm.stateAsyncProvider.stateSerializer.(*mock.MockCodec)

	stateChanges := []struct {
		stateName string
		value     string
		kind      ChangeKind
	}{
		{"stateNone", "valueNone", None},
		{"stateAdd", "valueAdd", Add},
		{"stateUpdate", "valueUpdate", Update},
		{"stateRemove", "valueRemove", Remove},
	}
	for _, sc := range stateChanges {
		sm.stateChangeTracker.Store(sc.stateName, &ChangeMetadata{Kind: sc.kind, Value: sc.value})
	}

	// 3 operations: 1 Add, 1 Update, 1 Remove
	mockClient.EXPECT().SaveStateTransactionally(ctx, sm.actorTypeName, sm.actorID, gomock.Len(3))
	// 2 times: 1 Add, 1 Update
	mockCodec.EXPECT().Marshal(gomock.Any()).Times(2)

	err := sm.Save()
	require.NoError(t, err)
	err = sm.stateManagerCtx.Save(ctx)
	require.NoError(t, err)
}

func TestFlush(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		context context.Context
	}{
		{"no context", nil},
		{"context", ctx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newMockStateManager(t)
			stateChanges := []struct {
				stateName string
				value     string
				kind      ChangeKind
			}{
				{"stateNone", "valueNone", None},
				{"stateAdd", "valueAdd", Add},
				{"stateUpdate", "valueUpdate", Update},
				{"stateRemove", "valueRemove", Remove},
			}
			for _, sc := range stateChanges {
				sm.stateChangeTracker.Store(sc.stateName, &ChangeMetadata{Kind: sc.kind, Value: sc.value})
			}

			if tt.context == nil {
				sm.Flush()
			} else {
				sm.stateManagerCtx.Flush(tt.context)
			}

			for _, sc := range stateChanges {
				val, ok := sm.stateChangeTracker.Load(sc.stateName)
				if sc.kind == Remove {
					assert.Nil(t, val)
					assert.False(t, ok)
				} else {
					metadata := val.(*ChangeMetadata)
					assert.True(t, ok)
					assert.Equal(t, None, metadata.Kind)
					assert.Equal(t, sc.value, metadata.Value)
				}
			}
		})
	}
}

func TestNewActorStateManager(t *testing.T) {
	type args struct {
		actorTypeName      string
		actorID            string
		stateAsyncProvider *DaprStateAsyncProvider
	}
	tests := []struct {
		name string
		args args
		want *stateManager
	}{
		{
			name: "init",
			args: args{
				actorTypeName:      "test",
				actorID:            "fn",
				stateAsyncProvider: &DaprStateAsyncProvider{},
			},
			want: &stateManager{
				&stateManagerCtx{
					actorTypeName:      "test",
					actorID:            "fn",
					stateAsyncProvider: &DaprStateAsyncProvider{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewActorStateManager(tt.args.actorTypeName, tt.args.actorID, tt.args.stateAsyncProvider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewActorStateManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewActorStateManagerContext(t *testing.T) {
	type args struct {
		actorTypeName      string
		actorID            string
		stateAsyncProvider *DaprStateAsyncProvider
	}
	tests := []struct {
		name string
		args args
		want *stateManagerCtx
	}{
		{
			name: "init",
			args: args{
				actorTypeName:      "test",
				actorID:            "fn",
				stateAsyncProvider: &DaprStateAsyncProvider{},
			},
			want: &stateManagerCtx{
				actorTypeName:      "test",
				actorID:            "fn",
				stateAsyncProvider: &DaprStateAsyncProvider{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewActorStateManagerContext(tt.args.actorTypeName, tt.args.actorID, tt.args.stateAsyncProvider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewActorStateManagerContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
