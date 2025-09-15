/*
Copyright 2024 The Dapr Authors
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
	"errors"
	"reflect"

	"google.golang.org/protobuf/types/known/structpb"

	"google.golang.org/protobuf/types/known/anypb"

	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

// conversationRequest object - currently unexported as used in a functions option pattern
// Deprecated: use ConversationRequestAlpha2 and ConverseAlpha2 instead.
type conversationRequest struct {
	name        string
	inputs      []ConversationInput
	Parameters  map[string]*anypb.Any
	Metadata    map[string]string
	ContextID   *string
	ScrubPII    *bool // Scrub PII from the output
	Temperature *float64
}

// NewConversationRequest defines a request with a component name and one or more inputs as a slice for the
// ConverseAlpha1 method.
// Deprecated: use ConversationRequestAlpha2 and ConverseAlpha2 instead.
func NewConversationRequest(llmName string, inputs []ConversationInput) conversationRequest {
	return conversationRequest{
		name:   llmName,
		inputs: inputs,
	}
}

// Deprecated: use the new ConversationRequestAlpha2 struct instead.
type conversationRequestOption func(request *conversationRequest)

// ConversationInput defines a single input for a conversation request for ConverseAlpha1.
// Deprecated: use ConversationInput in ConversationRequestAlpha2 instead.
type ConversationInput struct {
	// The content to send to the llm.
	Content string
	// The role of the message.
	Role *string
	// Whether to Scrub PII from the input
	ScrubPII *bool
}

// ConversationResponse is the basic response from a conversationRequest for ConverseAlpha1.
type ConversationResponse struct {
	ContextID string
	Outputs   []ConversationResult
}

// ConversationResult is the individual result from a conversation response for ConverseAlpha1.
type ConversationResult struct {
	Result     string
	Parameters map[string]*anypb.Any
}

// WithParameters should be used to provide parameters for custom fields.
// Deprecated: use ConversationRequestAlpha2.Parameters instead.
func WithParameters(parameters map[string]*anypb.Any) conversationRequestOption {
	return func(o *conversationRequest) {
		o.Parameters = parameters
	}
}

// WithMetadata used to define metadata to be passed to components.
// Deprecated: use ConversationRequestAlpha2.Metadata instead.
func WithMetadata(metadata map[string]string) conversationRequestOption {
	return func(o *conversationRequest) {
		o.Metadata = metadata
	}
}

// WithContextID to provide a new context or continue an existing one.
// Deprecated: use ConversationRequestAlpha2.ContextID instead.
func WithContextID(id string) conversationRequestOption {
	return func(o *conversationRequest) {
		o.ContextID = &id
	}
}

// WithScrubPII to define whether the outputs should have PII removed.
// Deprecated: use ConversationRequestAlpha2.ScrubPII instead.
func WithScrubPII(scrub bool) conversationRequestOption {
	return func(o *conversationRequest) {
		o.ScrubPII = &scrub
	}
}

// WithTemperature to specify which way the LLM leans.
// Deprecated: use ConversationRequestAlpha2.Temperature instead.
func WithTemperature(temp float64) conversationRequestOption {
	return func(o *conversationRequest) {
		o.Temperature = &temp
	}
}

// ConverseAlpha1 can invoke an LLM given a request created by the NewConversationRequest function.
// Deprecated: use ConverseAlpha2 instead.
func (c *GRPCClient) ConverseAlpha1(ctx context.Context, req conversationRequest, options ...conversationRequestOption) (*ConversationResponse, error) {
	//nolint:staticcheck
	cinputs := make([]*runtimev1pb.ConversationInput, len(req.inputs))
	for i, in := range req.inputs {
		//nolint:staticcheck
		cinputs[i] = &runtimev1pb.ConversationInput{
			Content:  in.Content,
			Role:     in.Role,
			ScrubPII: in.ScrubPII,
		}
	}

	for _, opt := range options {
		if opt != nil {
			opt(&req)
		}
	}
	//nolint:staticcheck
	request := runtimev1pb.ConversationRequest{
		Name:        req.name,
		ContextID:   req.ContextID,
		Inputs:      cinputs,
		Parameters:  req.Parameters,
		Metadata:    req.Metadata,
		ScrubPII:    req.ScrubPII,
		Temperature: req.Temperature,
	}

	resp, err := c.protoClient.ConverseAlpha1(ctx, &request)
	if err != nil {
		return nil, err
	}

	outputs := make([]ConversationResult, len(resp.GetOutputs()))
	for i, o := range resp.GetOutputs() {
		outputs[i] = ConversationResult{
			Result:     o.GetResult(),
			Parameters: o.GetParameters(),
		}
	}

	return &ConversationResponse{
		ContextID: resp.GetContextID(),
		Outputs:   outputs,
	}, nil
}

// ConversationToolsAlpha2 is one of the ConversationToolsFunctionAlpha2 types.
type ConversationToolsAlpha2 ConversationToolsFunctionAlpha2

func (ct *ConversationToolsAlpha2) toProto() []*runtimev1pb.ConversationTools {
	if ct != nil {
		protoTools := make([]*runtimev1pb.ConversationTools, 1)

		protoTools[0] = &runtimev1pb.ConversationTools{
			ToolTypes: &runtimev1pb.ConversationTools_Function{
				Function: &runtimev1pb.ConversationToolsFunction{
					Name:        ct.Name,
					Description: ct.Description,
					Parameters:  ct.Parameters,
				},
			},
		}

		return protoTools
	}

	return nil
}

type ConversationInputAlpha2 struct {
	Messages []*ConversationMessageAlpha2
	ScrubPII *bool
}

func (ci *ConversationInputAlpha2) toProto() *runtimev1pb.ConversationInputAlpha2 {
	if ci == nil {
		return nil
	}

	// validate messages
	for _, m := range ci.Messages {
		if m == nil {
			continue
		}
		if !m.Validate() {
			return nil
		}
	}

	if ci.Messages == nil || len(ci.Messages) == 0 {
		return nil
	}
	messages := make([]*runtimev1pb.ConversationMessage, len(ci.Messages))
	for i, m := range ci.Messages {
		if m == nil {
			messages[i] = nil
			continue
		}

		protoMsg, err := m.toProto()
		if err != nil {
			return nil
		}
		messages[i] = protoMsg
	}

	return &runtimev1pb.ConversationInputAlpha2{
		Messages: messages,
		ScrubPii: ci.ScrubPII,
	}
}

type ConversationToolsFunctionAlpha2 struct {
	Name        string
	Description *string
	Parameters  *structpb.Struct
}

type ConversationMessageAlpha2 struct {
	// oneof conversationmessagedeveloper, conversationmessagesystem, conversationmessageuser, conversationmessageassistant, conversationmessagetool
	ConversationMessageOfDeveloper *ConversationMessageOfDeveloperAlpha2
	ConversationMessageOfSystem    *ConversationMessageOfSystemAlpha2
	ConversationMessageOfUser      *ConversationMessageOfUserAlpha2
	ConversationMessageOfAssistant *ConversationMessageOfAssistantAlpha2
	ConversationMessageOfTool      *ConversationMessageOfToolAlpha2
}

// Validate ensures that exactly one of the oneof fields is set, not more, not less.
func (cm *ConversationMessageAlpha2) Validate() bool {
	if cm == nil {
		return false
	}

	fields := []interface{}{
		cm.ConversationMessageOfDeveloper,
		cm.ConversationMessageOfSystem,
		cm.ConversationMessageOfUser,
		cm.ConversationMessageOfAssistant,
		cm.ConversationMessageOfTool,
	}
	count := 0
	for _, f := range fields {
		if f != nil && !reflect.ValueOf(f).IsZero() {
			count++
		}
	}
	return count == 1
}

func (cm *ConversationMessageAlpha2) toProto() (*runtimev1pb.ConversationMessage, error) {
	if !cm.Validate() {
		return nil, errors.New("exactly one of the oneof fields must be set")
	}

	var protoMsg runtimev1pb.ConversationMessage

	switch {
	case cm.ConversationMessageOfDeveloper != nil:
		var content []*runtimev1pb.ConversationMessageContent
		if cm.ConversationMessageOfDeveloper.Content != nil {
			content = make([]*runtimev1pb.ConversationMessageContent, len(cm.ConversationMessageOfDeveloper.Content))
			for i, c := range cm.ConversationMessageOfDeveloper.Content {
				if c != nil && c.Text != nil {
					content[i] = &runtimev1pb.ConversationMessageContent{
						Text: *c.Text,
					}
				} else {
					content[i] = nil
				}
			}

			protoMsg.MessageTypes = &runtimev1pb.ConversationMessage_OfDeveloper{
				OfDeveloper: &runtimev1pb.ConversationMessageOfDeveloper{
					Name:    cm.ConversationMessageOfDeveloper.Name,
					Content: content,
				},
			}
		}
	case cm.ConversationMessageOfSystem != nil:
		var content []*runtimev1pb.ConversationMessageContent
		if cm.ConversationMessageOfSystem.Content != nil {
			content = make([]*runtimev1pb.ConversationMessageContent, len(cm.ConversationMessageOfSystem.Content))
			for i, c := range cm.ConversationMessageOfSystem.Content {
				if c != nil && c.Text != nil {
					content[i] = &runtimev1pb.ConversationMessageContent{
						Text: *c.Text,
					}
				} else {
					content[i] = nil
				}
			}
		}

		protoMsg.MessageTypes = &runtimev1pb.ConversationMessage_OfSystem{
			OfSystem: &runtimev1pb.ConversationMessageOfSystem{
				Name:    cm.ConversationMessageOfSystem.Name,
				Content: content,
			},
		}
	case cm.ConversationMessageOfUser != nil:
		var content []*runtimev1pb.ConversationMessageContent
		if cm.ConversationMessageOfUser.Content != nil {
			content = make([]*runtimev1pb.ConversationMessageContent, len(cm.ConversationMessageOfUser.Content))
			for i, c := range cm.ConversationMessageOfUser.Content {
				if c != nil && c.Text != nil {
					content[i] = &runtimev1pb.ConversationMessageContent{
						Text: *c.Text,
					}
				} else {
					content[i] = nil
				}
			}
		}

		protoMsg.MessageTypes = &runtimev1pb.ConversationMessage_OfUser{
			OfUser: &runtimev1pb.ConversationMessageOfUser{
				Name:    cm.ConversationMessageOfUser.Name,
				Content: content,
			},
		}
	case cm.ConversationMessageOfAssistant != nil:
		var content []*runtimev1pb.ConversationMessageContent
		if cm.ConversationMessageOfAssistant.Content != nil {
			content = make([]*runtimev1pb.ConversationMessageContent, len(cm.ConversationMessageOfAssistant.Content))
			for i, c := range cm.ConversationMessageOfAssistant.Content {
				if c != nil && c.Text != nil {
					content[i] = &runtimev1pb.ConversationMessageContent{
						Text: *c.Text,
					}
				} else {
					content[i] = nil
				}
			}
		}

		var toolCalls []*runtimev1pb.ConversationToolCalls
		if cm.ConversationMessageOfAssistant.ToolCalls != nil {
			toolCalls = make([]*runtimev1pb.ConversationToolCalls, len(cm.ConversationMessageOfAssistant.ToolCalls))
			for i, t := range cm.ConversationMessageOfAssistant.ToolCalls {
				if t != nil {
					toolCalls[i] = &runtimev1pb.ConversationToolCalls{
						Id: &t.ID,
						ToolTypes: &runtimev1pb.ConversationToolCalls_Function{
							Function: &runtimev1pb.ConversationToolCallsOfFunction{
								Name:      t.ToolTypes.Name,
								Arguments: t.ToolTypes.Arguments,
							},
						},
					}
				} else {
					toolCalls[i] = nil
				}
			}
		}

		protoMsg.MessageTypes = &runtimev1pb.ConversationMessage_OfAssistant{
			OfAssistant: &runtimev1pb.ConversationMessageOfAssistant{
				Name:      cm.ConversationMessageOfAssistant.Name,
				Content:   content,
				ToolCalls: toolCalls,
			},
		}
	case cm.ConversationMessageOfTool != nil:
		var content []*runtimev1pb.ConversationMessageContent
		if cm.ConversationMessageOfTool.Content != nil {
			content = make([]*runtimev1pb.ConversationMessageContent, len(cm.ConversationMessageOfTool.Content))
			for i, c := range cm.ConversationMessageOfTool.Content {
				if c != nil && c.Text != nil {
					content[i] = &runtimev1pb.ConversationMessageContent{
						Text: *c.Text,
					}
				} else {
					content[i] = nil
				}
			}
		}

		protoMsg.MessageTypes = &runtimev1pb.ConversationMessage_OfTool{
			OfTool: &runtimev1pb.ConversationMessageOfTool{
				ToolId:  cm.ConversationMessageOfTool.ToolID,
				Name:    *cm.ConversationMessageOfTool.Name,
				Content: content,
			},
		}
	}

	return &protoMsg, nil
}

type ConversationMessageContentAlpha2 struct {
	Text *string
}

type ConversationMessageOfDeveloperAlpha2 struct {
	Name    *string
	Content []*ConversationMessageContentAlpha2
}

type ConversationMessageOfSystemAlpha2 struct {
	Name    *string
	Content []*ConversationMessageContentAlpha2
}

type ConversationMessageOfUserAlpha2 struct {
	Name    *string
	Content []*ConversationMessageContentAlpha2
}

type ConversationMessageOfAssistantAlpha2 struct {
	Name      *string
	Content   []*ConversationMessageContentAlpha2
	ToolCalls []*ConversationToolCallsAlpha2
}

type ConversationMessageOfToolAlpha2 struct {
	ToolID  *string
	Name    *string
	Content []*ConversationMessageContentAlpha2
}

type ConversationRequestAlpha2 struct {
	Name        string // LLM component name
	ContextID   *string
	Inputs      []*ConversationInputAlpha2
	Parameters  map[string]*anypb.Any
	Metadata    map[string]string
	ScrubPII    *bool // Scrub PII from the output
	Temperature *float64
	Tools       []*ConversationToolsAlpha2
	ToolChoice  *ToolChoiceAlpha2
}

type ConversationResponseAlpha2 struct {
	ContextID string
	Outputs   []*ConversationResultAlpha2
}

type ConversationResultAlpha2 struct {
	Choices []*ConversationResultChoicesAlpha2
}

type ConversationResultChoicesAlpha2 struct {
	FinishReason string
	Index        int64
	Message      *ConversationResultMessageAlpha2
}

type ConversationResultMessageAlpha2 struct {
	Content   string
	ToolCalls []*ConversationToolCallsAlpha2
}

type ConversationToolCallsAlpha2 struct {
	ID        string
	ToolTypes ConversationToolAlpha2
}

type ConversationToolAlpha2 struct {
	Name      string
	Arguments string
}

// ToolChoiceAlpha2 defines how to handle tools in a conversation request.
// It can be either none, auto, required or a specific tool name (with a custom string).
type ToolChoiceAlpha2 string

// convert string to *string
func (tc *ToolChoiceAlpha2) toPtr() *string {
	if tc == nil {
		return nil
	}
	s := string(*tc)
	return &s
}

const (
	ToolChoiceNoneAlpha2     ToolChoiceAlpha2 = "none"
	ToolChoiceAutoAlpha2     ToolChoiceAlpha2 = "auto"
	ToolChoiceRequiredAlpha2 ToolChoiceAlpha2 = "required"
)

type conversationRequestOptionAlpha2 func(request *ConversationRequestAlpha2)

func (c *GRPCClient) ConverseAlpha2(ctx context.Context, request ConversationRequestAlpha2, options ...conversationRequestOptionAlpha2) (*ConversationResponseAlpha2, error) {
	var inputs []*runtimev1pb.ConversationInputAlpha2
	if request.Inputs != nil {
		inputs = make([]*runtimev1pb.ConversationInputAlpha2, len(request.Inputs))
		for i, in := range request.Inputs {
			if in.Messages == nil {
				inputs[i] = nil
				continue
			}
			protoInput := in.toProto()
			if protoInput == nil {
				return nil, errors.New("failed to convert ConversationInputAlpha2 to proto")
			}
			inputs[i] = protoInput
		}
	}

	var tools []*runtimev1pb.ConversationTools
	if request.Tools != nil {
		tools = make([]*runtimev1pb.ConversationTools, 0, len(request.Tools))
		for _, t := range request.Tools {
			if t == nil {
				continue
			}
			protoTools := t.toProto()
			if protoTools == nil {
				return nil, errors.New("failed to convert ConversationToolsAlpha2 to proto")
			}
			tools = append(tools, protoTools...)
		}
	}

	req := runtimev1pb.ConversationRequestAlpha2{
		Name:        request.Name,
		ContextId:   request.ContextID,
		Inputs:      inputs,
		Parameters:  request.Parameters,
		Metadata:    request.Metadata,
		ScrubPii:    request.ScrubPII,
		Temperature: request.Temperature,
		Tools:       tools,
		ToolChoice:  request.ToolChoice.toPtr(),
	}

	resp, err := c.protoClient.ConverseAlpha2(ctx, &req)
	if err != nil {
		return nil, err
	}

	outputs := make([]*ConversationResultAlpha2, len(resp.GetOutputs()))
	for i, o := range resp.GetOutputs() {
		choices := make([]*ConversationResultChoicesAlpha2, len(o.GetChoices()))
		for j, c := range o.GetChoices() {
			toolCalls := make([]*ConversationToolCallsAlpha2, len(c.GetMessage().GetToolCalls()))
			for k, t := range c.GetMessage().GetToolCalls() {
				toolCalls[k] = &ConversationToolCallsAlpha2{
					ID: t.GetId(),
					ToolTypes: ConversationToolAlpha2{
						Name:      t.GetFunction().GetName(),
						Arguments: t.GetFunction().GetArguments(),
					},
				}
			}

			choices[j] = &ConversationResultChoicesAlpha2{
				FinishReason: c.GetFinishReason(),
				Index:        c.GetIndex(),
				Message: &ConversationResultMessageAlpha2{
					Content:   c.GetMessage().GetContent(),
					ToolCalls: toolCalls,
				},
			}
		}
		outputs[i] = &ConversationResultAlpha2{
			Choices: choices,
		}
	}

	response := &ConversationResponseAlpha2{
		ContextID: resp.GetContextId(),
		Outputs:   outputs,
	}

	return response, nil
}
