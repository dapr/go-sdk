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
)

// go test -timeout 30s ./client -count 1 -run ^TestGetSecret$
func TestGetSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("without store", func(t *testing.T) {
		out, err := testClient.GetSecret(ctx, "", "key1", nil)
		assert.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("without key", func(t *testing.T) {
		out, err := testClient.GetSecret(ctx, "store", "", nil)
		assert.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("without meta", func(t *testing.T) {
		out, err := testClient.GetSecret(ctx, "store", "key1", nil)
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})

	t.Run("with meta", func(t *testing.T) {
		in := map[string]string{"k1": "v1", "k2": "v2"}
		out, err := testClient.GetSecret(ctx, "store", "key1", in)
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})
}

func TestGetBulkSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("without store", func(t *testing.T) {
		out, err := testClient.GetBulkSecret(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("without meta", func(t *testing.T) {
		out, err := testClient.GetBulkSecret(ctx, "store", nil)
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})

	t.Run("with meta", func(t *testing.T) {
		in := map[string]string{"k1": "v1", "k2": "v2"}
		out, err := testClient.GetBulkSecret(ctx, "store", in)
		assert.Nil(t, err)
		assert.NotNil(t, out)
	})
}
