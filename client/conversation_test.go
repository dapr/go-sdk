/*
Copyright 2026 The Dapr Authors
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
	"testing"

	"github.com/dapr/kit/ptr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestNewConversationRequest(t *testing.T) {
	tests := []struct {
		name           string
		llmName        string
		inputs         []ConversationInput
		wantName       string
		wantInputCount int
	}{
		{
			name:           "empty inputs",
			llmName:        "my-llm",
			inputs:         nil,
			wantName:       "my-llm",
			wantInputCount: 0,
		},
		{
			name:           "one input",
			llmName:        "openai",
			inputs:         []ConversationInput{{Content: "hello", Role: ptr.Of("user")}},
			wantName:       "openai",
			wantInputCount: 1,
		},
		{
			name:           "many inputs",
			llmName:        "llm",
			inputs:         []ConversationInput{{Content: "a"}, {Content: "b"}},
			wantName:       "llm",
			wantInputCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewConversationRequest(tt.llmName, tt.inputs)
			assert.Equal(t, tt.wantName, req.name)
			assert.Len(t, req.inputs, tt.wantInputCount)
		})
	}
}

func TestConversationMessageAlpha2Validate(t *testing.T) {
	userMsg := &ConversationMessageOfUserAlpha2{
		Name:    ptr.Of("user"),
		Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("hi")}},
	}
	systemMsg := &ConversationMessageOfSystemAlpha2{
		Name:    ptr.Of("system"),
		Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("you are helpful")}},
	}
	developerMsg := &ConversationMessageOfDeveloperAlpha2{
		Name:    ptr.Of("dev"),
		Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("instruction")}},
	}
	assistantMsg := &ConversationMessageOfAssistantAlpha2{
		Name:    ptr.Of("assistant"),
		Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("response")}},
		ToolCalls: []*ConversationToolCallsAlpha2{
			{ID: "call-1", ToolTypes: ConversationToolAlpha2{Name: "get_weather", Arguments: `{"location":"NYC"}`}},
		},
	}
	toolMsg := &ConversationMessageOfToolAlpha2{
		ToolID:  ptr.Of("call-1"),
		Name:    ptr.Of("get_weather"),
		Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("sunny")}},
	}

	tests := []struct {
		name    string
		msg     *ConversationMessageAlpha2
		isValid bool
	}{
		{
			name:    "nil message",
			msg:     nil,
			isValid: false,
		},
		{
			name:    "empty message",
			msg:     &ConversationMessageAlpha2{},
			isValid: false,
		},
		{
			name: "one user message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfUser: userMsg,
			},
			isValid: true,
		},
		{
			name: "one system message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfSystem: systemMsg,
			},
			isValid: true,
		},
		{
			name: "one developer message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfDeveloper: developerMsg,
			},
			isValid: true,
		},
		{
			name: "one assistant message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfAssistant: assistantMsg,
			},
			isValid: true,
		},
		{
			name: "one tool message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfTool: toolMsg,
			},
			isValid: true,
		},
		{
			name: "many messages",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfUser:   userMsg,
				ConversationMessageOfSystem: systemMsg,
			},
			isValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.Validate()
			assert.Equal(t, tt.isValid, got)
		})
	}
}

func TestConversationMessageAlpha2ToProto(t *testing.T) {
	tests := []struct {
		name    string
		msg     *ConversationMessageAlpha2
		isValid bool
	}{
		{
			name:    "empty message",
			msg:     &ConversationMessageAlpha2{},
			isValid: true,
		},
		{
			name: "user message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfUser: &ConversationMessageOfUserAlpha2{
					Name:    ptr.Of("user"),
					Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("hello")}},
				},
			},
			isValid: false,
		},
		{
			name: "system message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfSystem: &ConversationMessageOfSystemAlpha2{
					Name:    ptr.Of("system"),
					Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("helper")}},
				},
			},
			isValid: false,
		},
		{
			name: "developer message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfDeveloper: &ConversationMessageOfDeveloperAlpha2{
					Name:    ptr.Of("dev"),
					Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("instruction")}},
				},
			},
			isValid: false,
		},
		{
			name: "assistant message with content",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfAssistant: &ConversationMessageOfAssistantAlpha2{
					Name:      ptr.Of("assistant"),
					Content:   []*ConversationMessageContentAlpha2{{Text: ptr.Of("here is the result")}},
					ToolCalls: nil,
				},
			},
			isValid: false,
		},
		{
			name: "assistant message with tool calls",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfAssistant: &ConversationMessageOfAssistantAlpha2{
					Name:    ptr.Of("assistant"),
					Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("calling tool")}},
					ToolCalls: []*ConversationToolCallsAlpha2{
						{ID: "call-1", ToolTypes: ConversationToolAlpha2{Name: "get_weather", Arguments: `{"loc":"Boston"}`}},
					},
				},
			},
			isValid: false,
		},
		{
			name: "tool message",
			msg: &ConversationMessageAlpha2{
				ConversationMessageOfTool: &ConversationMessageOfToolAlpha2{
					ToolID:  ptr.Of("call-1"),
					Name:    ptr.Of("get_weather"),
					Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("72")}},
				},
			},
			isValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.msg.toProto()
			if tt.isValid {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func TestConversationToolsAlpha2ToProto(t *testing.T) {
	desc := "a test tool"
	tests := []struct {
		name      string
		tools     *ConversationToolsAlpha2
		wantCount int
	}{
		{
			name:      "nil tools",
			tools:     nil,
			wantCount: 0,
		},
		{
			name: "valid tool",
			tools: &ConversationToolsAlpha2{
				Name:        "my_tool",
				Description: &desc,
				Parameters:  nil,
			},
			wantCount: 1,
		},
		{
			name: "valid tool with empty description",
			tools: &ConversationToolsAlpha2{
				Name:        "simple",
				Description: nil,
				Parameters:  nil,
			},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tools.toProto()
			if tt.wantCount == 0 {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Len(t, got, tt.wantCount)
			assert.Equal(t, tt.tools.Name, got[0].GetFunction().GetName())
		})
	}
}

func TestConversationInputAlpha2ToProto(t *testing.T) {
	validUserMsg := &ConversationMessageAlpha2{
		ConversationMessageOfUser: &ConversationMessageOfUserAlpha2{
			Name:    ptr.Of("user"),
			Content: []*ConversationMessageContentAlpha2{{Text: ptr.Of("hi")}},
		},
	}

	tests := []struct {
		name      string
		input     *ConversationInputAlpha2
		wantCount int
	}{
		{
			name:      "nil input",
			input:     nil,
			wantCount: 0,
		},
		{
			name: "nil messages",
			input: &ConversationInputAlpha2{
				Messages: nil,
			},
			wantCount: 0,
		},
		{
			name: "empty messages",
			input: &ConversationInputAlpha2{
				Messages: []*ConversationMessageAlpha2{},
			},
			wantCount: 0,
		},
		{
			name: "invalid message in list",
			input: &ConversationInputAlpha2{
				Messages: []*ConversationMessageAlpha2{{}},
			},
			wantCount: 0,
		},
		{
			name: "valid single message",
			input: &ConversationInputAlpha2{
				Messages: []*ConversationMessageAlpha2{validUserMsg},
			},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.toProto()
			if tt.wantCount == 0 {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Len(t, got.GetMessages(), 1)
		})
	}
}

func TestToolChoiceAlpha2ToPtr(t *testing.T) {
	custom := ToolChoiceAlpha2("my_tool")
	tests := []struct {
		name       string
		toolChoice *ToolChoiceAlpha2
		want       string
	}{
		{
			name:       "nil",
			toolChoice: nil,
		},
		{
			name:       "none",
			toolChoice: func() *ToolChoiceAlpha2 { v := ToolChoiceNoneAlpha2; return &v }(),
			want:       "none",
		},
		{
			name:       "auto",
			toolChoice: func() *ToolChoiceAlpha2 { v := ToolChoiceAutoAlpha2; return &v }(),
			want:       "auto",
		},
		{
			name:       "required",
			toolChoice: func() *ToolChoiceAlpha2 { v := ToolChoiceRequiredAlpha2; return &v }(),
			want:       "required",
		},
		{
			name:       "custom",
			toolChoice: &custom,
			want:       "my_tool",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.toolChoice.toPtr()
			if tt.toolChoice == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestConverseAlpha2(t *testing.T) {
	ctx := t.Context()
	client := &GRPCClient{protoClient: nil}
	req := ConversationRequestAlpha2{
		Name: "test-llm",
		Inputs: []*ConversationInputAlpha2{
			{
				Messages: []*ConversationMessageAlpha2{
					{},
				},
			},
		},
	}

	_, err := client.ConverseAlpha2(ctx, req)
	// expect error because protoInput is nil
	require.Error(t, err)
}

func TestConversationRequestOptions(t *testing.T) {
	req := NewConversationRequest("llm", []ConversationInput{{Content: "x"}})
	ctxID := "ctx-1"
	scrub := true
	temp := 0.5

	WithContextID(ctxID)(&req)
	WithScrubPII(scrub)(&req)
	WithTemperature(temp)(&req)
	WithParameters(map[string]*anypb.Any{
		"key": {
			Value: []byte("value"),
		},
	})(&req)
	WithMetadata(map[string]string{
		"key": "value",
	})(&req)

	assert.NotNil(t, req.ContextID)
	assert.Equal(t, ctxID, *req.ContextID)
	assert.NotNil(t, req.ScrubPII)
	assert.True(t, *req.ScrubPII)
	assert.NotNil(t, req.Temperature)
	assert.InDelta(t, temp, *req.Temperature, 1e-9)
	assert.NotNil(t, req.Parameters)
	assert.Equal(t, map[string]*anypb.Any{
		"key": {
			Value: []byte("value"),
		},
	}, req.Parameters)
	assert.NotNil(t, req.Metadata)
	assert.Equal(t, map[string]string{
		"key": "value",
	}, req.Metadata)
}
