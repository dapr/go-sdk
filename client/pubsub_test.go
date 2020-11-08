package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type _testStructwithText struct {
	key1, key2 string
}

type _testStructwithTextandNumbers struct {
	key1 string
	key2 int
}

type _testStructwithSlices struct {
	key1 []string
	key2 []int
}

// go test -timeout 30s ./client -count 1 -run ^TestPublishEvent$
func TestPublishEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("with data", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messagebus", "test", []byte("ping"))
		assert.Nil(t, err)
	})

	t.Run("without data", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messagebus", "test", nil)
		assert.Nil(t, err)
	})

	t.Run("with empty topic name", func(t *testing.T) {
		err := testClient.PublishEvent(ctx, "messagebus", "", []byte("ping"))
		assert.NotNil(t, err)
	})

	t.Run("from struct with text", func(t *testing.T) {
		testdata := _testStructwithText{
			key1: "value1",
			key2: "value2",
		}
		err := testClient.PublishEventfromStruct(ctx, "messagebus", "test", testdata)
		assert.Nil(t, err)
	})

	t.Run("from struct with text and numbers", func(t *testing.T) {
		testdata := _testStructwithTextandNumbers{
			key1: "value1",
			key2: 2500,
		}
		err := testClient.PublishEventfromStruct(ctx, "messagebus", "test", testdata)
		assert.Nil(t, err)
	})

	t.Run("from struct with slices", func(t *testing.T) {
		testdata := _testStructwithSlices{
			key1: []string{"value1", "value2", "value3"},
			key2: []int{25, 40, 600},
		}
		err := testClient.PublishEventfromStruct(ctx, "messagebus", "test", testdata)
		assert.Nil(t, err)
	})
}
