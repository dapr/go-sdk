package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dapr/go-sdk/service/common"
	"github.com/stretchr/testify/assert"
)

func TestBindingHandlerWithoutData(t *testing.T) {
	t.Parallel()

	s := newService("")
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
	t.Parallel()

	data := `{"name": "test"}`
	s := newService("")
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