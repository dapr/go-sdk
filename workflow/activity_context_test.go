package workflow

import (
	"context"
	"encoding/json"
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
}
