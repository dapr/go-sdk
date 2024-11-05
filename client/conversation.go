/*
Copyright 2022 The Dapr Authors
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
	"fmt"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

// ConversationRequest is the conversation request object.
type ConversationRequest struct {
	// The ID of an existing chat (like in ChatGPT)
	ContextID string
	// Inputs for the conversation, support multiple input in one time.
	Inputs []ConversationInput
	// Parameters for all custom fields.
	Parameters map[string]*anypb.Any
	// The metadata passing to conversation components.
	Metadata map[string]string
	// Scrub PII data that comes back from the LLM
	ScrubPII bool
	// Temperature for the LLM to optimize for creativity or predictability
	Temperature *float64
}

// ConversationInput is a an input to an LLM conversation
type ConversationInput struct {
	// The message to send to the llm
	Message string
	// The role to set for the message
	Role string
	// Scrub PII data that goes into the LLM
	ScrubPII bool
}

// ConversationResponse is the conversation response object.
type ConversationResponse struct {
	// The ID of an existing chat (like in ChatGPT)
	ContextID string
	// An array of results.
	Outputs []ConversationResult
}

// ConversationResult is an output for a conversation request input
type ConversationResult struct {
	// Result for the conversation input.
	Result string
	// Parameters for all custom fields.
	Parameters map[string]*anypb.Any
}

// ConverseAlpha1 issues a prompt to an LLM provider.
func (c *GRPCClient) ConverseAlpha1(ctx context.Context, llmName string, request *ConversationRequest) (*ConversationResponse, error) {
	if llmName == "" {
		return nil, errors.New("llmName is empty")
	}

	if request == nil {
		return nil, errors.New("request is nil")
	}

	if len(request.Inputs) == 0 {
		return nil, errors.New("conversation inputs must contain at least one item")
	}

	req := pb.ConversationRequest{
		Name:        llmName,
		Metadata:    request.Metadata,
		ScrubPII:    &request.ScrubPII,
		Temperature: request.Temperature,
		Parameters:  request.Parameters,
		Inputs:      []*pb.ConversationInput{},
	}
	if request.ContextID != "" {
		req.ContextID = &request.ContextID
	}

	for _, i := range request.Inputs {
		req.Inputs = append(req.Inputs, &pb.ConversationInput{
			Message:  i.Message,
			Role:     &i.Role,
			ScrubPII: &i.ScrubPII,
		})
	}

	resp, err := c.protoClient.ConverseAlpha1(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("error prompting llm: %w", err)
	}

	cv := &ConversationResponse{}
	if resp.ContextID != nil {
		cv.ContextID = *resp.ContextID
	}

	for _, o := range resp.Outputs {
		cv.Outputs = append(cv.Outputs, ConversationResult{
			Result:     o.Result,
			Parameters: o.Parameters,
		})
	}

	return cv, nil
}
