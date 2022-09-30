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

	"github.com/stretchr/testify/assert"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

const (
	testLockStore = "store"
)

func TestLock(t *testing.T) {
	ctx := context.Background()

	t.Run("try lock invalid store name", func(t *testing.T) {
		r, err := testClient.TryLockAlpha1(ctx, "", &LockRequest{})
		assert.Nil(t, r)
		assert.Error(t, err)
	})

	t.Run("try lock invalid request", func(t *testing.T) {
		r, err := testClient.TryLockAlpha1(ctx, testLockStore, nil)
		assert.Nil(t, r)
		assert.Error(t, err)
	})

	t.Run("try lock", func(t *testing.T) {
		r, err := testClient.TryLockAlpha1(ctx, testLockStore, &LockRequest{
			LockOwner:       "owner1",
			ResourceID:      "resource1",
			ExpiryInSeconds: 5,
		})
		assert.NotNil(t, r)
		assert.NoError(t, err)
		assert.True(t, r.Success)
	})

	t.Run("unlock invalid store name", func(t *testing.T) {
		r, err := testClient.UnlockAlpha1(ctx, "", &UnlockRequest{
			LockOwner:  "owner1",
			ResourceID: "resource1",
		})
		assert.Nil(t, r)
		assert.Error(t, err)
	})

	t.Run("unlock invalid request", func(t *testing.T) {
		r, err := testClient.UnlockAlpha1(ctx, "testLockStore", nil)
		assert.Nil(t, r)
		assert.Error(t, err)
	})

	t.Run("unlock", func(t *testing.T) {
		r, err := testClient.UnlockAlpha1(ctx, testLockStore, &UnlockRequest{
			LockOwner:  "owner1",
			ResourceID: "resource1",
		})
		assert.NotNil(t, r)
		assert.NoError(t, err)
		assert.Equal(t, pb.UnlockResponse_SUCCESS.String(), r.Status)
	})
}
