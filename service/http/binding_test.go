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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dapr/go-sdk/service/common"
)

func TestBindingHandlerWithoutHandler(t *testing.T) {
	s := newServer("", nil)
	err := s.AddBindingInvocationHandler("/", nil)
	assert.Errorf(t, err, "expected error adding nil binding event handler")
}

func TestBindingHandlerWithoutData(t *testing.T) {
	s := newServer("", nil)
	err := s.AddBindingInvocationHandler("/", func(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
		if in == nil {
			return nil, errors.New("nil input")
		}
		if in.Data != nil {
			return nil, errors.New("invalid input data")
		}
		return nil, nil
	})
	assert.NoErrorf(t, err, "error adding binding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "{}", resp.Body.String())
}

func TestBindingHandlerWithData(t *testing.T) {
	data := `{"name": "test"}`
	s := newServer("", nil)
	err := s.AddBindingInvocationHandler("/", func(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
		if in == nil {
			return nil, errors.New("nil input")
		}
		return []byte("test"), nil
	})
	assert.NoErrorf(t, err, "error adding binding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "test", resp.Body.String())
}

func bindingHandlerFn(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	return []byte("test"), nil
}

func bindingHandlerFnWithError(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	return nil, errors.New("intentional error")
}

func TestBindingHandlerErrors(t *testing.T) {
	data := `{"name": "test"}`
	s := newServer("", nil)
	err := s.AddBindingInvocationHandler("", bindingHandlerFn)
	assert.Errorf(t, err, "expected error adding binding event handler sans route")

	err = s.AddBindingInvocationHandler("errors", bindingHandlerFnWithError)
	assert.NoErrorf(t, err, "error adding binding event handler sans slash")

	req, err := http.NewRequest(http.MethodPost, "/errors", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}
