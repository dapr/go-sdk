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
		failedEntries := testClient.PublishEvents(ctx, "", "test", []interface{}{"ping", "pong"})
		assert.Len(t, failedEntries, 2)
		// TODO: assert the data
	})

	t.Run("without topic name", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "", []interface{}{"ping", "pong"})
		assert.Len(t, failedEntries, 2)
		// TODO: assert the data
	})

	t.Run("with data", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"})
		assert.Len(t, failedEntries, 0)
	})

	t.Run("without data", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", nil)
		assert.Len(t, failedEntries, 0)
	})

	t.Run("from struct with text", func(t *testing.T) {
		testdata := _testStructwithText{
			Key1: "value1",
			Key2: "value2",
		}
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{testdata})
		assert.Len(t, failedEntries, 0)
	})

	t.Run("from struct with text and numbers", func(t *testing.T) {
		testdata := _testStructwithTextandNumbers{
			Key1: "value1",
			Key2: 2500,
		}
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{testdata})
		assert.Len(t, failedEntries, 0)
	})

	t.Run("from struct with slices", func(t *testing.T) {
		testdata := _testStructwithSlices{
			Key1: []string{"value1", "value2", "value3"},
			Key2: []int{25, 40, 600},
		}
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{testdata})
		assert.Len(t, failedEntries, 0)
	})

	t.Run("error serializing one event", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{make(chan struct{}), "pong"})
		assert.Len(t, failedEntries, 1)
		assert.Equal(t, failedEntries[0].Event, PublishEventsEvent{})
	})

	t.Run("with raw payload", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"}, PublishEventsWithRawPayload())
		assert.Len(t, failedEntries, 0)
	})

	t.Run("with metadata", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"}, PublishEventsWithMetadata(map[string]string{"key": "value"}))
		assert.Len(t, failedEntries, 0)
	})

	t.Run("with custom content type", func(t *testing.T) {
		failedEntries := testClient.PublishEvents(ctx, "messages", "test", []interface{}{"ping", "pong"}, PublishEventsWithContentType("text/plain"))
		assert.Len(t, failedEntries, 0)
	})
}
