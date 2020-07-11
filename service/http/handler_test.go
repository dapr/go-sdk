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

	mux := http.NewServeMux()
	s, err := NewService(mux)
	assert.NoErrorf(t, err, "error creating service")

	err = s.AddTopicEventHandler("test", "/", func(ctx context.Context, e TopicEvent) error {
		if e.DataContentType != "application/json" {
			t.Fatalf("invalid data content type: %s", e.DataContentType)
		}
		return nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	err = s.HandleSubscriptions()
	assert.NoErrorf(t, err, "error handling subscriptions")

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
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestInvocationHandler(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	s, err := NewService(mux)
	assert.NoErrorf(t, err, "error creating service")

	data := `{ "message": "pong" }`
	contentType := "application/json"
	err = s.AddInvocationHandler("/", func(ctx context.Context, in *InvocationEvent) (out []byte, err error) {
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
	mux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "error reading response body")
	assert.Equal(t, data, string(b))
}
