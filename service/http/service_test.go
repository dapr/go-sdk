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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStoppingUnstartedService(t *testing.T) {
	s := newServer("", nil)
	assert.NotNil(t, s)
	err := s.Stop()
	assert.NoError(t, err)
}

func TestStoppingStartedService(t *testing.T) {
	s := newServer(":3333", nil)
	assert.NotNil(t, s)

	go func() {
		if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	// Wait for the server to start
	time.Sleep(200 * time.Millisecond)
	assert.NoError(t, s.Stop())
}

func TestStartingStoppedService(t *testing.T) {
	s := newServer(":3333", nil)
	assert.NotNil(t, s)
	stopErr := s.Stop()
	assert.NoError(t, stopErr)

	startErr := s.Start()
	assert.Error(t, startErr, "expected starting a stopped server to raise an error")
	assert.Equal(t, startErr.Error(), http.ErrServerClosed.Error())
}

func TestSettingOptions(t *testing.T) {
	req, err := http.NewRequest(http.MethodOptions, "/", nil)
	assert.NoErrorf(t, err, "error creating request")
	w := httptest.NewRecorder()
	setOptions(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.NotNil(t, resp)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST,OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "authorization, origin, content-type, accept", resp.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "POST,OPTIONS", resp.Header.Get("Allow"))
}

func testRequest(t *testing.T, s *Server, r *http.Request, expectedStatusCode int) {
	t.Helper()

	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, r)
	resp := rr.Result()
	defer resp.Body.Close()
	assert.NotNil(t, resp)
	assert.Equal(t, expectedStatusCode, resp.StatusCode)
}

func testRequestWithResponseBody(t *testing.T, s *Server, r *http.Request, expectedStatusCode int, expectedBody []byte) {
	t.Helper()

	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, r)
	rez := rr.Result()
	defer rez.Body.Close()
	rspBody, err := io.ReadAll(rez.Body)
	assert.Nil(t, err)
	assert.NotNil(t, rez)
	assert.Equal(t, expectedStatusCode, rez.StatusCode)
	assert.Equal(t, expectedBody, rspBody)
}
