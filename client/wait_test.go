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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGrpcWait(t *testing.T) {
	ctx := context.Background()

	// Clean up env. var just in case
	os.Setenv(clientTimoutSecondsEnvVarName, "")
	_, err := getClientTimeoutSeconds()
	assert.NoError(t, err)

	t.Run("Happy Case Client test", func(t *testing.T) {
		err := testClient.Wait(ctx, 5*time.Second)
		assert.NoError(t, err)
	})

	t.Run("Waiting after shutdown fails as there is nothing to wait for", func(t *testing.T) {
		testClient.Shutdown(ctx)
		err := testClient.Wait(ctx, 5*time.Second)

		assert.Error(t, err, "Waiting after shutdown should fail as there is no connection left")
		assert.Equal(t, errWaitTimedOut, err)
	})

	t.Run("No wait just doesn't work because there is always a delay to accept connections", func(t *testing.T) {
		err := testClient.Wait(ctx, 0*time.Second)
		assert.Error(t, err)
		assert.Equal(t, errWaitTimedOut, err)
	})
}
