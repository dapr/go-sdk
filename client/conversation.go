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
	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

type conversationRequestOptions struct {
	Parameters  map[string]*anypb.Any
	Metadata    map[string]string
	ContextID   *string
	ScrubPII    *bool // Scrub PII from the output
	Temperature *float64
}

type conversationRequestOption func(request *conversationRequestOptions)

type ConversationInput struct {
	Message  string
	Role     *string
	ScrubPII *bool // Scrub PII from the input
}

type ConversationResponse struct {
	ContextID string
	Outputs   []ConversationResult
}

type ConversationResult struct {
	Result     string
	Parameters map[string]*anypb.Any
}

func WithParameters(parameters map[string]*anypb.Any) conversationRequestOption {
	return func(o *conversationRequestOptions) {
		o.Parameters = parameters
	}
}

func WithMetadata(metadata map[string]string) conversationRequestOption {
	return func(o *conversationRequestOptions) {
		o.Metadata = metadata
	}
}

func WithContextID(id string) conversationRequestOption {
	return func(o *conversationRequestOptions) {
		o.ContextID = &id
	}
}

func WithScrubPII(scrub bool) conversationRequestOption {
	return func(o *conversationRequestOptions) {
		o.ScrubPII = &scrub
	}
}

func WithTemperature(temp float64) conversationRequestOption {
	return func(o *conversationRequestOptions) {
		o.Temperature = &temp
	}
}

func (c *GRPCClient) ConverseAlpha1(ctx context.Context, componentName string, inputs []ConversationInput, options ...conversationRequestOption) (*ConversationResponse, error) {

	var cinputs []*runtimev1pb.ConversationInput
	for _, i := range inputs {
		cinputs = append(cinputs, &runtimev1pb.ConversationInput{
			Message:  i.Message,
			Role:     i.Role,
			ScrubPII: i.ScrubPII,
		})
	}

	var o conversationRequestOptions
	for _, opt := range options {
		if opt != nil {
			opt(&o)
		}
	}

	request := runtimev1pb.ConversationRequest{
		Name:        componentName,
		ContextID:   o.ContextID,
		Inputs:      cinputs,
		Parameters:  o.Parameters,
		Metadata:    o.Metadata,
		ScrubPII:    o.ScrubPII,
		Temperature: o.Temperature,
	}

	resp, err := c.protoClient.ConverseAlpha1(ctx, &request)
	if err != nil {
		return nil, err
	}

	var outputs []ConversationResult
	for _, i := range resp.GetOutputs() {
		outputs = append(outputs, ConversationResult{
			Result:     i.GetResult(),
			Parameters: i.GetParameters(),
		})
	}

	return &ConversationResponse{
		ContextID: resp.GetContextID(),
		Outputs:   outputs,
	}, nil
}
