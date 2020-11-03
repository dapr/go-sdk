package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/dapr/go-sdk/service/common"
	"github.com/stretchr/testify/assert"
)

func testTopicFunc(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	if e == nil {
		return false, errors.New("nil content")
	}
	if e.DataContentType != "application/json" {
		return false, fmt.Errorf("invalid content type: %s", e.DataContentType)
	}
	return false, nil
}

func testErrorTopicFunc(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	return true, errors.New("error to cause a retry")
}

func TestEventNilHandler(t *testing.T) {
	s := newServer("", nil)
	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
		Route:      "/",
		Metadata:   map[string]string{},
	}
	err := s.AddTopicEventHandler(sub, nil)
	assert.Errorf(t, err, "expected error adding event handler")
}

func TestEventHandler(t *testing.T) {
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

	s := newServer("", nil)

	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "test",
		Route:      "/",
		Metadata:   map[string]string{},
	}
	err := s.AddTopicEventHandler(sub, testTopicFunc)
	assert.NoErrorf(t, err, "error adding event handler")

	sub2 := &common.Subscription{
		PubsubName: "messages",
		Topic:      "errors",
		Route:      "/errors",
		Metadata:   map[string]string{},
	}
	err = s.AddTopicEventHandler(sub2, testErrorTopicFunc)
	assert.NoErrorf(t, err, "error adding error event handler")

	s.registerSubscribeHandler()

	makeEventRequest(t, s, "/", data, http.StatusOK)
	makeEventRequest(t, s, "/", "", http.StatusSeeOther)
	makeEventRequest(t, s, "/", "not JSON", http.StatusSeeOther)
	makeEventRequest(t, s, "/errors", data, http.StatusOK)
}

func makeEventRequest(t *testing.T, s *Server, route, data string, expectedStatusCode int) {
	req, err := http.NewRequest(http.MethodPost, route, strings.NewReader(data))
	assert.NoErrorf(t, err, "error creating request: %s", data)
	req.Header.Set("Content-Type", "application/json")
	testRequest(t, s, req, expectedStatusCode)
}

func TestAddingInvalidEventHandlers(t *testing.T) {
	s := newServer("", nil)
	err := s.AddTopicEventHandler(nil, testTopicFunc)
	assert.Errorf(t, err, "expected error adding no sub event handler")

	sub := &common.Subscription{Metadata: map[string]string{}}
	err = s.AddTopicEventHandler(sub, testTopicFunc)
	assert.Errorf(t, err, "expected error adding empty sub event handler")

	sub.Topic = "test"
	err = s.AddTopicEventHandler(sub, testTopicFunc)
	assert.Errorf(t, err, "expected error adding sub without component event handler")

	sub.PubsubName = "messages"
	err = s.AddTopicEventHandler(sub, testTopicFunc)
	assert.Errorf(t, err, "expected error adding sub without route event handler")
}
