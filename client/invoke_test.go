package client

import (
	"context"
	"testing"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	"github.com/stretchr/testify/assert"
)

type _testStructwithText struct {
	Key1, Key2 string
}

type _testStructwithTextandNumbers struct {
	Key1 string
	Key2 int
}

type _testStructwithSlices struct {
	Key1 []string
	Key2 []int
}

// go test -timeout 30s ./client -count 1 -run ^TestInvokeServiceWithContent$
func TestInvokeServiceWithContent(t *testing.T) {
	ctx := context.Background()
	data := "ping"

	t.Run("with content", func(t *testing.T) {
		content := &DataContent{
			ContentType: "text/plain",
			Data:        []byte(data),
		}
		resp, err := testClient.InvokeServiceWithContent(ctx, "test", "fn", "post", content)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, string(resp), data)

	})

	t.Run("without content", func(t *testing.T) {
		resp, err := testClient.InvokeService(ctx, "test", "fn", "get")
		assert.Nil(t, err)
		assert.Nil(t, resp)

	})

	t.Run("without service ID", func(t *testing.T) {
		_, err := testClient.InvokeService(ctx, "", "fn", "get")
		assert.NotNil(t, err)
	})
	t.Run("without method", func(t *testing.T) {
		_, err := testClient.InvokeService(ctx, "test", "", "get")
		assert.NotNil(t, err)
	})
	t.Run("without verb", func(t *testing.T) {
		_, err := testClient.InvokeService(ctx, "test", "fn", "")
		assert.NotNil(t, err)
	})
	t.Run("from struct with text", func(t *testing.T) {
		testdata := _testCustomContentwithText{
			Key1: "value1",
			Key2: "value2",
		}
		_, err := testClient.InvokeServiceWithCustomContent(ctx, "test", "fn", "post", "text/plain", testdata)
		assert.Nil(t, err)
	})

	t.Run("from struct with text and numbers", func(t *testing.T) {
		testdata := _testCustomContentwithTextandNumbers{
			Key1: "value1",
			Key2: 2500,
		}
		_, err := testClient.InvokeServiceWithCustomContent(ctx, "test", "fn", "post", "text/plain", testdata)
		assert.Nil(t, err)
	})

	t.Run("from struct with slices", func(t *testing.T) {
		testdata := _testCustomContentwithSlices{
			Key1: []string{"value1", "value2", "value3"},
			Key2: []int{25, 40, 600},
		}
		_, err := testClient.InvokeServiceWithCustomContent(ctx, "test", "fn", "post", "text/plain", testdata)
		assert.Nil(t, err)
	})
}

func TestVerbParsing(t *testing.T) {
	t.Run("valid lower case", func(t *testing.T) {
		v := verbToHTTPExtension("post")
		assert.NotNil(t, v)
		assert.Equal(t, v1.HTTPExtension_POST, v.Verb)
	})

	t.Run("valid upper case", func(t *testing.T) {
		v := verbToHTTPExtension("GET")
		assert.NotNil(t, v)
		assert.Equal(t, v1.HTTPExtension_GET, v.Verb)
	})

	t.Run("invalid verb", func(t *testing.T) {
		v := verbToHTTPExtension("BAD")
		assert.NotNil(t, v)
		assert.Equal(t, v1.HTTPExtension_NONE, v.Verb)
	})
}
