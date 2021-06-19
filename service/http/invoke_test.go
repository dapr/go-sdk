package http

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dapr/go-sdk/service/common"
	"github.com/stretchr/testify/assert"
)

func TestInvocationHandlerWithoutHandler(t *testing.T) {
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/", nil)
	assert.Errorf(t, err, "expected error adding event handler")
}

func TestInvocationHandlerWithData(t *testing.T) {
	data := `{"name": "test", "data": hellow}`
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/", func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
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
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "error reading response body")
	assert.Equal(t, data, string(b))
}

func TestInvocationHandlerWithoutInputData(t *testing.T) {
	s := newServer("", nil)
	err := s.AddServiceInvocationHandler("/", func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
		if in == nil || in.Data != nil {
			err = errors.New("nil input")
			return
		}
		return &common.Content{}, nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "error reading response body")
	assert.NotNil(t, b)
	assert.Equal(t, "", string(b))
}

func emptyInvocationFn(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	return nil, nil
}

func TestInvocationHandlerWithInvalidRoute(t *testing.T) {
	s := newServer("", nil)

	err := s.AddServiceInvocationHandler("no-slash", emptyInvocationFn)
	assert.NoErrorf(t, err, "error adding no slash route event handler")

	err = s.AddServiceInvocationHandler("", emptyInvocationFn)
	assert.Errorf(t, err, "expected error from adding no route event handler")

	err = s.AddServiceInvocationHandler("/a", emptyInvocationFn)
	assert.NoErrorf(t, err, "error adding event handler")

	makeEventRequest(t, s, "/b", "", http.StatusNotFound)
}

func errorInvocationFn(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	return nil, errors.New("intentional test error")
}

func TestInvocationHandlerWithError(t *testing.T) {
	s := newServer("", nil)

	err := s.AddServiceInvocationHandler("/error", errorInvocationFn)
	assert.NoErrorf(t, err, "error adding error event handler")

	makeEventRequest(t, s, "/error", "", http.StatusInternalServerError)
}
