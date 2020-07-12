package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	data := `{ "message": "pong" }`
	contentType := "application/json"

	s := newService()
	err := s.AddInvocationHandler("/", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
		if in == nil {
			t.Fatal("nil invocation events")
		}
		if in.ContentType == contentType {
			return []byte(data), nil
		}
		return []byte("test"), nil
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

func TestInvocationHandlerWithoutData(t *testing.T) {
	t.Parallel()

	data := "test"
	s := newService()
	err := s.AddInvocationHandler("/", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
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

func TestInvocationHandlerWithInvalidRoute(t *testing.T) {
	t.Parallel()

	s := newService()
	err := s.AddInvocationHandler("/a", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
		return []byte("test"), nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/b", nil)
	assert.NoErrorf(t, err, "error creating request")

	resp := httptest.NewRecorder()
	s.Mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}
