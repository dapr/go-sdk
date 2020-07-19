package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventHandler(t *testing.T) {
	t.Parallel()

	data := `{
		"specversion" : "1.0",
		"type" : "com.github.pull.create",
		"source" : "https://github.com/cloudevents/spec/pull",
		"subject" : "123",
		"id" : "A234-1234-1234",
		"time" : "2018-04-05T17:31:00Z",
		"comexampleextension1" : "value",
		"comexampleothervalue" : 5,
		"datacontenttype" : "application/json",
		"data" : "eyJtZXNzYWdlIjoiaGVsbG8ifQ=="
	}`

	s := newService("")

	err := s.AddTopicEventHandler("test", "/", func(ctx context.Context, e *TopicEvent) error {
		if e.DataContentType != "application/json" {
			t.Fatalf("invalid data content type: %s", e.DataContentType)
		}
		return nil
	})
	assert.NoErrorf(t, err, "error adding event handler")

	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.registerSubscribeHandler()
	s.mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
