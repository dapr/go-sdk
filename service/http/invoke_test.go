package http

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dapr/go-sdk/service"
	"github.com/stretchr/testify/assert"
)

func TestInvocationHandlerWithData(t *testing.T) {
	t.Parallel()

	data := fmt.Sprintf(`{
		"name": "test",
		"data": %s
	}`, []byte("hellow"))

	s := newService("")
	err := s.AddServiceInvocationHandler("test", func(ctx context.Context, in *service.InvocationEvent) (out *service.InvocationEvent, err error) {
		if in == nil {
			err = errors.New("nil input")
			return
		}
		return in, nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(data))
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
	t.Parallel()

	data := `{
		"name": "test",
	}`

	s := newService("")
	err := s.AddServiceInvocationHandler("test", func(ctx context.Context, in *service.InvocationEvent) (out *service.InvocationEvent, err error) {
		if in == nil {
			err = errors.New("nil input")
			return
		}
		return &service.InvocationEvent{}, nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(data))
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

func TestInvocationHandlerWithInvalidRoute(t *testing.T) {
	t.Parallel()

	s := newService("")
	err := s.AddServiceInvocationHandler("/a", func(ctx context.Context, in *service.InvocationEvent) (out *service.InvocationEvent, err error) {
		return nil, nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/b", nil)
	assert.NoErrorf(t, err, "error creating request")

	resp := httptest.NewRecorder()
	s.mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}
