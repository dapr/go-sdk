/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type _testCustomContentwithText struct {
	Key1, Key2 string
}

type _testCustomContentwithTextandNumbers struct {
	Key1 string
	Key2 int
}

type _testCustomContentwithSlices struct {
	Key1 []string
	Key2 []int
}

// go test -timeout 30s ./client -count 1 -run ^TestPublishEvent$
func TestPublishEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("with data", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messages", "test", []byte("ping"))
		assert.Nil(t, err)
	})

	t.Run("without data", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messages", "test", nil)
		assert.Nil(t, err)
	})

	t.Run("with empty topic name", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messages", "", []byte("ping"))
		assert.NotNil(t, err)
	})

	t.Run("from struct with text", func(t *testing.T) {
		testdata := _testStructwithText{
			Key1: "value1",
			Key2: "value2",
		}
		err := testClient.PublishEventfromCustomContent(ctx, "messages", "test", testdata)
		assert.Nil(t, err)
	})

	t.Run("from struct with text and numbers", func(t *testing.T) {
		testdata := _testStructwithTextandNumbers{
			Key1: "value1",
			Key2: 2500,
		}
		err := testClient.PublishEventfromCustomContent(ctx, "messages", "test", testdata)
		assert.Nil(t, err)
	})

	t.Run("from struct with slices", func(t *testing.T) {
		testdata := _testStructwithSlices{
			Key1: []string{"value1", "value2", "value3"},
			Key2: []int{25, 40, 600},
		}
		err := testClient.PublishEventfromCustomContent(ctx, "messages", "test", testdata)
		assert.Nil(t, err)
	})

	t.Run("error serializing JSON", func(t *testing.T) {
		err := testClient.PublishEventfromCustomContent(ctx, "messages", "test", make(chan struct{}))
		assert.Error(t, err)
	})

	t.Run("raw payload", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messages", "test", []byte("ping"), PublishEventWithRawPayload())
		assert.Nil(t, err)
	})
}

// go test -timeout 30s ./client -count 1 -run ^TestPublishEvents$
func TestPublishEvents(t *testing.T) {
	ctx := context.Background()

	t.Run("without pubsub name", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "", "test", []interface{}{"ping", "pong"})
		assert.Error(t, res.Error)
		assert.Len(t, res.FailedEvents, 2)
		assert.Contains(t, res.FailedEvents, "ping")
		assert.Contains(t, res.FailedEvents, "pong")
	})

	t.Run("without topic name", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "", []interface{}{"ping", "pong"})
		assert.Error(t, res.Error)
		assert.Len(t, res.FailedEvents, 2)
		assert.Contains(t, res.FailedEvents, "ping")
		assert.Contains(t, res.FailedEvents, "pong")
	})

	t.Run("with data", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"})
		assert.Nil(t, res.Error)
		assert.Len(t, res.FailedEvents, 0)
	})

	t.Run("without data", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", nil)
		assert.Nil(t, res.Error)
		assert.Len(t, res.FailedEvents, 0)
	})

	t.Run("with struct data", func(t *testing.T) {
		testcases := []struct {
			name string
			data interface{}
		}{
			{
				name: "with text",
				data: _testStructwithText{
					Key1: "value1",
					Key2: "value2",
				},
			},
			{
				name: "with text and numbers",
				data: _testStructwithTextandNumbers{
					Key1: "value1",
					Key2: 2500,
				},
			},
			{
				name: "with slices",
				data: _testStructwithSlices{
					Key1: []string{"value1", "value2", "value3"},
					Key2: []int{25, 40, 600},
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{tc.data})
				assert.Nil(t, res.Error)
				assert.Len(t, res.FailedEvents, 0)
			})
		}
	})

	t.Run("error serializing one event", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{make(chan struct{}), "pong"})
		assert.Error(t, res.Error)
		assert.Len(t, res.FailedEvents, 1)
		assert.IsType(t, make(chan struct{}), res.FailedEvents[0])
	})

	t.Run("with raw payload", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"}, PublishEventsWithRawPayload())
		assert.Nil(t, res.Error)
		assert.Len(t, res.FailedEvents, 0)
	})

	t.Run("with metadata", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"}, PublishEventsWithMetadata(map[string]string{"key": "value"}))
		assert.Nil(t, res.Error)
		assert.Len(t, res.FailedEvents, 0)
	})

	t.Run("with custom content type", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"}, PublishEventsWithContentType("text/plain"))
		assert.Nil(t, res.Error)
		assert.Len(t, res.FailedEvents, 0)
	})

	t.Run("with events that will fail some events", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong", "fail-ping"})
		assert.Error(t, res.Error)
		assert.Len(t, res.FailedEvents, 1)
		assert.Contains(t, res.FailedEvents, "fail-ping")
	})

	t.Run("with events that will fail the entire request", func(t *testing.T) {
		res := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong", "failall-ping"})
		assert.Error(t, res.Error)
		assert.Len(t, res.FailedEvents, 3)
		assert.Contains(t, res.FailedEvents, "ping")
		assert.Contains(t, res.FailedEvents, "pong")
		assert.Contains(t, res.FailedEvents, "failall-ping")
	})
}

func TestCreateBulkPublishRequestEntry(t *testing.T) {
	type _testJSONStruct struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}

	type _testCloudEventStruct struct {
		ID          string `json:"id"`
		Source      string `json:"source"`
		SpecVersion string `json:"specversion"`
		Type        string `json:"type"`
		Data        string `json:"data"`
	}

	t.Run("should serialize and set content type", func(t *testing.T) {
		testcases := []struct {
			name                string
			data                interface{}
			expectedEvent       []byte
			expectedContentType string
			expectedError       bool
		}{
			{
				name:                "plain text",
				data:                "ping",
				expectedEvent:       []byte(`ping`),
				expectedContentType: "text/plain",
				expectedError:       false,
			},
			{
				name:                "raw bytes",
				data:                []byte("ping"),
				expectedEvent:       []byte(`ping`),
				expectedContentType: "application/octet-stream",
				expectedError:       false,
			},
			{
				name: "valid json",
				data: _testJSONStruct{
					Key1: "value1",
					Key2: "value2",
				},
				expectedEvent:       []byte(`{"key1":"value1","key2":"value2"}`),
				expectedContentType: "application/json",
				expectedError:       false,
			},
			{
				name: "valid cloudevent",
				data: _testCloudEventStruct{
					ID:          "123",
					Source:      "test",
					SpecVersion: "1.0",
					Type:        "test",
					Data:        "foo",
				},
				expectedEvent:       []byte(`{"id":"123","source":"test","specversion":"1.0","type":"test","data":"foo"}`),
				expectedContentType: "application/cloudevents+json",
				expectedError:       false,
			},
			{
				name:                "invalid json",
				data:                make(chan struct{}),
				expectedEvent:       nil,
				expectedContentType: "",
				expectedError:       true,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				entry, err := createBulkPublishRequestEntry(tc.data)
				if tc.expectedError {
					assert.Error(t, err)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, tc.expectedEvent, entry.Event)
					assert.Equal(t, tc.expectedContentType, entry.ContentType)
				}
			})
		}
	})

	t.Run("should set same entryID and metadata when provided", func(t *testing.T) {
		entry, err := createBulkPublishRequestEntry(PublishEventsEvent{
			ContentType: "text/plain",
			Data:        []byte("ping"),
			EntryID:     "123",
			Metadata:    map[string]string{"key": "value"},
		})
		assert.Nil(t, err)
		assert.Equal(t, "123", entry.EntryId)
		assert.Equal(t, map[string]string{"key": "value"}, entry.Metadata)
	})

	t.Run("should set random uuid as entryID when not provided", func(t *testing.T) {
		testcases := []struct {
			name string
			data interface{}
		}{
			{
				name: "plain text",
				data: "ping",
			},
			{
				name: "PublishEventsEvent",
				data: PublishEventsEvent{
					ContentType: "text/plain",
					Data:        []byte("ping"),
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				entry, err := createBulkPublishRequestEntry(tc.data)
				assert.Nil(t, err)
				assert.NotEmpty(t, entry.EntryId)
				assert.Nil(t, entry.Metadata)

				_, err = uuid.Parse(entry.EntryId)
				assert.Nil(t, err)
			})
		}
	})
}
