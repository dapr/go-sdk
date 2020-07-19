package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dapr/go-sdk/service"
	"github.com/stretchr/testify/assert"
)

func TestEventHandler(t *testing.T) {
	t.Parallel()

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

	s := newService("")

	err := s.AddTopicEventHandler("test", func(ctx context.Context, e *service.TopicEvent) error {
		if e.DataContentType != "application/json" {
			t.Fatalf("invalid data content type: %s", e.DataContentType)
		}
		return nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")

	rr := httptest.NewRecorder()
	s.registerSubscribeHandler()
	s.mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
