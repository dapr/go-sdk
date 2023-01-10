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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dapr/go-sdk/service/common"
)

func TestInvocationHandlerWithoutHandler(t *testing.T) {
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/hello", nil)
	assert.Errorf(t, err, "expected error adding event handler")

	err = s.AddServiceInvocationHandler("/", nil)
	assert.Errorf(t, err, "expected error adding event handler, invalid router")
}

func TestInvocationHandlerWithToken(t *testing.T) {
	data := `{"name": "test", "data": hello}`
	t.Setenv(common.AppAPITokenEnvVar, "app-dapr-token")
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/hello", func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
		if in == nil || in.Data == nil || in.ContentType == "" {
			err = errors.New("nil input")
			return
		}
		out = &common.Content{
			Data:        in.Data,
			ContentType: in.ContentType,
			DataTypeURL: in.DataTypeURL,
		}
		return
	})
	assert.NoErrorf(t, err, "adding event handler success")

	// forbbiden.
	req, err := http.NewRequest(http.MethodPost, "/hello", strings.NewReader(data))
	assert.NoErrorf(t, err, "creating request success")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNonAuthoritativeInfo, resp.Code)

	// pass.
	req.Header.Set(common.APITokenKey, os.Getenv(common.AppAPITokenEnvVar))
	resp = httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	_ = os.Unsetenv(common.AppAPITokenEnvVar)
}

func TestInvocationHandlerWithData(t *testing.T) {
	data := `{"name": "test", "data": hello}`
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/hello", func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
		if in == nil || in.Data == nil || in.ContentType == "" {
			err = errors.New("nil input")
			return
		}
		out = &common.Content{
			Data:        in.Data,
			ContentType: in.ContentType,
			DataTypeURL: in.DataTypeURL,
		}
		return
	})
	assert.NoErrorf(t, err, "adding event handler success")

	req, err := http.NewRequest(http.MethodPost, "/hello", strings.NewReader(data))
	assert.NoErrorf(t, err, "creating request success")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := io.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "reading response body success")
	assert.Equal(t, data, string(b))
}

func TestInvocationHandlerWithoutInputData(t *testing.T) {
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/hello", func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
		if in == nil || in.Data != nil {
			err = errors.New("nil input")
			return
		}
		return &common.Content{}, nil
	})
	assert.NoErrorf(t, err, "adding event handler success")

	req, err := http.NewRequest(http.MethodPost, "/hello", nil)
	assert.NoErrorf(t, err, "creating request success")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := io.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "reading response body success")
	assert.NotNil(t, b)
	assert.Equal(t, "", string(b))
}

func emptyInvocationFn(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	return nil, nil
}

func TestInvocationHandlerWithInvalidRoute(t *testing.T) {
	s := newServer("", nil)

	err := s.AddServiceInvocationHandler("no-slash", emptyInvocationFn)
	assert.NoErrorf(t, err, "adding no slash route event handler success")

	err = s.AddServiceInvocationHandler("", emptyInvocationFn)
	assert.Errorf(t, err, "expected error from adding no route event handler")

	err = s.AddServiceInvocationHandler("/a", emptyInvocationFn)
	assert.NoErrorf(t, err, "adding event handler success")

	makeEventRequest(t, s, "/b", "", http.StatusNotFound)
}

func errorInvocationFn(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	return nil, errors.New("intentional test error")
}

func TestInvocationHandlerWithError(t *testing.T) {
	s := newServer("", nil)

	err := s.AddServiceInvocationHandler("/error", errorInvocationFn)
	assert.NoErrorf(t, err, "adding error event handler success")

	makeEventRequest(t, s, "/error", "", http.StatusInternalServerError)
}
