package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	s, err := NewServer(mux)
	assert.NoErrorf(t, err, "error creating server")

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
