package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testingTaskActivityContext struct {
	inputBytes []byte
}

func (t *testingTaskActivityContext) GetInput(v any) error {
	return json.Unmarshal(t.inputBytes, &v)
}

func (t *testingTaskActivityContext) Context() context.Context {
	return context.TODO()
}

func TestActivityContext(t *testing.T) {
	inputString := "testInputString"
	inputBytes, err := json.Marshal(inputString)
	require.NoErrorf(t, err, "required no error, but got %v", err)

	ac := ActivityContext{ctx: &testingTaskActivityContext{inputBytes: inputBytes}}
	t.Run("test getinput", func(t *testing.T) {
		var inputReturn string
		err := ac.GetInput(&inputReturn)
		require.NoError(t, err)
		assert.Equal(t, inputString, inputReturn)
	})

	t.Run("test context", func(t *testing.T) {
		assert.Equal(t, context.TODO(), ac.Context())
	})
}

func TestMarshalData(t *testing.T) {
	t.Run("test nil input", func(t *testing.T) {
		out, err := marshalData(nil)
		require.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("test string input", func(t *testing.T) {
		out, err := marshalData("testString")
		require.NoError(t, err)
		fmt.Println(out)
		assert.Equal(t, []byte{0x22, 0x74, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22}, out)
	})
}
