package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/stretchr/testify/assert"
)

func TestEventHandler(t *testing.T) {
	t.Parallel()

	s := newService()
	err := s.AddTopicEventHandler("test", "/", func(ctx context.Context, e TopicEvent) error {
		if e.DataContentType != "application/json" {
			t.Fatalf("invalid data content type: %s", e.DataContentType)
		}
		return nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	data := `{
		"specversion" : "0.3",
		"type" : "io.dapr.test",
		"source" : "https://dapr.io/test",
		"subject" : "test",
		"id" : "A123-4567-8910",
		"time" : "2020-07-11T17:31:00Z",
		"datacontenttype" : "application/json",
		"data" : "{\"message\": \"hi\"}"
	}`

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")

	rr := httptest.NewRecorder()
	s.registerSubscribeHandler()
	s.Mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestInvocationHandlerWithData(t *testing.T) {
	t.Parallel()

	data := `{
		"v1" : "test",
		"v2" : 1
	}`
	contentType := "application/json"

	s := newService()
	err := s.AddInvocationEventHandler("/", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
		if in == nil {
			err = errors.New("nil input")
			return
		}
		return in.Data, nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", contentType)

	resp := httptest.NewRecorder()
	s.Mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "error reading response body")
	assert.Equal(t, data, string(b))
}

func TestInvocationHandlerWithoutInputData(t *testing.T) {
	t.Parallel()

	data := "test"
	s := newService()
	err := s.AddInvocationEventHandler("/", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
		if in != nil {
			return nil, errors.New("expected nil input")
		}
		return []byte(data), nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoErrorf(t, err, "error creating request")

	resp := httptest.NewRecorder()
	s.Mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "error reading response body")
	assert.NotNil(t, b)
	assert.Equal(t, data, string(b))
}

func TestInvocationHandlerWithoutOutputData(t *testing.T) {
	t.Parallel()

	s := newService()
	err := s.AddInvocationEventHandler("/", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
		if in != nil {
			return nil, errors.New("expected nil input")
		}
		return nil, nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoErrorf(t, err, "error creating request")

	resp := httptest.NewRecorder()
	s.Mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "error reading response body")
	assert.NotNil(t, b)
	assert.Equal(t, 0, len(b))
}

func TestInvocationHandlerWithInvalidRoute(t *testing.T) {
	t.Parallel()

	s := newService()
	err := s.AddInvocationEventHandler("/a", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
		return []byte("test"), nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/b", nil)
	assert.NoErrorf(t, err, "error creating request")

	resp := httptest.NewRecorder()
	s.Mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestBindingHandlerWithoutData(t *testing.T) {
	t.Parallel()

	e := &pb.BindingEventRequest{}
	b, err := json.Marshal(e)
	assert.NoErrorf(t, err, "error serializing binding event")

	s := newService()
	err = s.AddBindingEventHandler("/", func(ctx context.Context, in *BindingEvent) error {
		if in == nil {
			return errors.New("nil input")
		}
		if in.Data != nil {
			return errors.New("invalid input data")
		}
		return nil
	})
	assert.NoErrorf(t, err, "error adding binding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	assert.NoErrorf(t, err, "error creating request")

	resp := httptest.NewRecorder()
	s.Mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
