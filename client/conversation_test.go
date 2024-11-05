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

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const (
	testLLM = "llm"
)

func TestConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("conversation missing inputs", func(t *testing.T) {
		r, err := testClient.ConverseAlpha1(ctx, testLLM, &ConversationRequest{})
		assert.Nil(t, r)
		require.Error(t, err)
	})

	t.Run("conversation missing llm name", func(t *testing.T) {
		r, err := testClient.ConverseAlpha1(ctx, "", &ConversationRequest{Inputs: []ConversationInput{{}}})
		assert.Nil(t, r)
		require.Error(t, err)
	})

	t.Run("conversation nil request", func(t *testing.T) {
		r, err := testClient.ConverseAlpha1(ctx, testLLM, nil)
		assert.Nil(t, r)
		require.Error(t, err)
	})
}
