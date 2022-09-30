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

package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandlerWithoutHandler(t *testing.T) {
	s := newServer("", nil)
	err := s.AddHealthCheckHandler("/", nil)
	assert.Errorf(t, err, "expected error adding nil health check handler")
}

func TestHealthCheckHandler(t *testing.T) {
	t.Run("health check with http status 200", func(t *testing.T) {
		s := newServer("", nil)
		err := s.AddHealthCheckHandler("/", func(ctx context.Context) (err error) {
			return nil
		})

		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoErrorf(t, err, "error creating request")
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		s.mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("health check with http status 500", func(t *testing.T) {
		s := newServer("", nil)
		err := s.AddHealthCheckHandler("/", func(ctx context.Context) (err error) {
			fmt.Println("hello,owrl")
			return errors.New("app is unhealthy")
		})

		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoErrorf(t, err, "error creating request")
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		s.mux.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}
