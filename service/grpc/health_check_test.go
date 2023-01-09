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

package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHealthCheckHandler(ctx context.Context) (err error) {
	return nil
}

func testHealthCheckHandlerWithError(ctx context.Context) (err error) {
	return errors.New("app is unhealthy")
}

func TestHealthCheckHandlerForErrors(t *testing.T) {
	server := getTestServer()
	err := server.AddHealthCheckHandler("", nil)
	assert.Errorf(t, err, "expected error on nil health check handler")
}

// go test -timeout 30s ./service/grpc -count 1 -run ^TestHealthCheck$
func TestHealthCheck(t *testing.T) {
	ctx := context.Background()

	server := getTestServer()
	startTestServer(server)

	t.Run("health check without handler", func(t *testing.T) {
		_, err := server.HealthCheck(ctx, nil)
		assert.Error(t, err)
	})

	err := server.AddHealthCheckHandler("", testHealthCheckHandler)
	assert.Nil(t, err)

	t.Run("health check with handler", func(t *testing.T) {
		_, err = server.HealthCheck(ctx, nil)
		assert.Nil(t, err)
	})

	err = server.AddHealthCheckHandler("", testHealthCheckHandlerWithError)
	assert.Nil(t, err)

	t.Run("health check with error handler", func(t *testing.T) {
		_, err = server.HealthCheck(ctx, nil)
		assert.Error(t, err)
	})

	stopTestServer(t, server)
}
