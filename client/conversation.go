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

	"google.golang.org/protobuf/types/known/anypb"

	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

// conversationRequest object - currently unexported as used in a functions option pattern
type conversationRequest struct {
	name        string
	inputs      []ConversationInput
	Parameters  map[string]*anypb.Any
	Metadata    map[string]string
	ContextID   *string
	ScrubPII    *bool // Scrub PII from the output
	Temperature *float64
}

// NewConversationRequest defines a request with a component name and one or more inputs as a slice
func NewConversationRequest(llmName string, inputs []ConversationInput) conversationRequest {
	return conversationRequest{
		name:   llmName,
		inputs: inputs,
	}
}

type conversationRequestOption func(request *conversationRequest)

// ConversationInput defines a single input.
type ConversationInput struct {
	// The content to send to the llm.
	Content string
	// The role of the message.
	Role *string
	// Whether to Scrub PII from the input
	ScrubPII *bool
}

// ConversationResponse is the basic response from a conversationRequest.
type ConversationResponse struct {
	ContextID string
	Outputs   []ConversationResult
}

// ConversationResult is the individual
type ConversationResult struct {
	Result     string
	Parameters map[string]*anypb.Any
}

// WithParameters should be used to provide parameters for custom fields.
func WithParameters(parameters map[string]*anypb.Any) conversationRequestOption {
	return func(o *conversationRequest) {
		o.Parameters = parameters
	}
}

// WithMetadata used to define metadata to be passed to components.
func WithMetadata(metadata map[string]string) conversationRequestOption {
	return func(o *conversationRequest) {
		o.Metadata = metadata
	}
}

// WithContextID to provide a new context or continue an existing one.
func WithContextID(id string) conversationRequestOption {
	return func(o *conversationRequest) {
		o.ContextID = &id
	}
}

// WithScrubPII to define whether the outputs should have PII removed.
func WithScrubPII(scrub bool) conversationRequestOption {
	return func(o *conversationRequest) {
		o.ScrubPII = &scrub
	}
}

// WithTemperature to specify which way the LLM leans.
func WithTemperature(temp float64) conversationRequestOption {
	return func(o *conversationRequest) {
		o.Temperature = &temp
	}
}

// ConverseAlpha1 can invoke an LLM given a request created by the NewConversationRequest function.
func (c *GRPCClient) ConverseAlpha1(ctx context.Context, req conversationRequest, options ...conversationRequestOption) (*ConversationResponse, error) {
	cinputs := make([]*runtimev1pb.ConversationInput, len(req.inputs))
	for i, in := range req.inputs {
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
