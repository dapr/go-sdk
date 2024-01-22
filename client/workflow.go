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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

type StartWorkflowRequest struct {
	InstanceID        string // Optional instance identifier
	WorkflowComponent string
	WorkflowName      string
	Options           map[string]string // Optional metadata
	Input             any               // Optional input
	SendRawInput      bool              // Set to True in order to disable serialization on the input
}

type StartWorkflowResponse struct {
	InstanceID string
}

type GetWorkflowRequest struct {
	InstanceID        string
	WorkflowComponent string
}

type GetWorkflowResponse struct {
	InstanceID    string
	WorkflowName  string
	CreatedAt     time.Time
	LastUpdatedAt time.Time
	RuntimeStatus string
	Properties    map[string]string
}

type PurgeWorkflowRequest struct {
	InstanceID        string
	WorkflowComponent string
}

type TerminateWorkflowRequest struct {
	InstanceID        string
	WorkflowComponent string
}

type PauseWorkflowRequest struct {
	InstanceID        string
	WorkflowComponent string
}

type ResumeWorkflowRequest struct {
	InstanceID        string
	WorkflowComponent string
}

type RaiseEventWorkflowRequest struct {
	InstanceID        string
	WorkflowComponent string
	EventName         string
	EventData         any
	SendRawData       bool // Set to True in order to disable serialization on the data
}

// StartWorkflowBeta1 starts a workflow using the beta1 spec.
func (c *GRPCClient) StartWorkflowBeta1(ctx context.Context, req *StartWorkflowRequest) (*StartWorkflowResponse, error) {
	if req.InstanceID == "" {
		req.InstanceID = uuid.New().String()
	}
	if req.WorkflowComponent == "" {
		return nil, errors.New("failed to start workflow: WorkflowComponent must be supplied")
	}
	if req.WorkflowName == "" {
		return nil, errors.New("failed to start workflow: WorkflowName must be supplied")
	}

	var input []byte
	var err error
	if (!req.SendRawInput) && (req.Input != nil) {
		input, err = json.Marshal(req.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input: %v", err)
		}
	} else {
		input = []byte(fmt.Sprintf("%v", req.Input))
	}

	resp, err := c.protoClient.StartWorkflowBeta1(c.withAuthToken(ctx), &pb.StartWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
		WorkflowName:      req.WorkflowName,
		Options:           req.Options,
		Input:             input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start workflow instance: %v", err)
	}
	return &StartWorkflowResponse{
		InstanceID: resp.GetInstanceId(),
	}, nil
}

// GetWorkflowBeta1 gets the status of a workflow using the beta1 spec.
func (c *GRPCClient) GetWorkflowBeta1(ctx context.Context, req *GetWorkflowRequest) (*GetWorkflowResponse, error) {
	if req.InstanceID == "" {
		return nil, errors.New("failed to get workflow status: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return nil, errors.New("failed to get workflow status: WorkflowComponent must be supplied")
	}
	resp, err := c.protoClient.GetWorkflowBeta1(c.withAuthToken(ctx), &pb.GetWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %v", err)
	}
	if resp.GetCreatedAt() == nil {
		resp.CreatedAt = timestamppb.Now()
	}
	if resp.GetLastUpdatedAt() == nil {
		resp.LastUpdatedAt = timestamppb.Now()
	}
	return &GetWorkflowResponse{
		InstanceID:    resp.GetInstanceId(),
		WorkflowName:  resp.GetWorkflowName(),
		CreatedAt:     resp.GetCreatedAt().AsTime(),
		LastUpdatedAt: resp.GetLastUpdatedAt().AsTime(),
		RuntimeStatus: resp.GetRuntimeStatus(),
		Properties:    resp.GetProperties(),
	}, nil
}

// PurgeWorkflowBeta1 removes all metadata relating to a specific workflow using the beta1 spec.
func (c *GRPCClient) PurgeWorkflowBeta1(ctx context.Context, req *PurgeWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to purge workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to purge workflow: WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.PurgeWorkflowBeta1(c.withAuthToken(ctx), &pb.PurgeWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to purge workflow: %v", err)
	}
	return nil
}

// TerminateWorkflowBeta1 stops a workflow using the beta1 spec.
func (c *GRPCClient) TerminateWorkflowBeta1(ctx context.Context, req *TerminateWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to terminate workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to terminate workflow, WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.TerminateWorkflowBeta1(ctx, &pb.TerminateWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate workflow: %v", err)
	}
	return nil
}

// PauseWorkflowBeta1 pauses a workflow that can be resumed later using the beta1 spec.
func (c *GRPCClient) PauseWorkflowBeta1(ctx context.Context, req *PauseWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to pause workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to pause workflow, WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.PauseWorkflowBeta1(ctx, &pb.PauseWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to pause workflow: %v", err)
	}
	return nil
}

// ResumeWorkflowBeta1 resumes a paused workflow using the beta1 spec.
func (c *GRPCClient) ResumeWorkflowBeta1(ctx context.Context, req *ResumeWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to resume workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to resume workflow: WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.ResumeWorkflowBeta1(c.withAuthToken(ctx), &pb.ResumeWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to resume workflow: %v", err)
	}
	return nil
}

// RaiseEventWorkflowBeta1 raises an event on a workflow using the beta1 spec.
func (c *GRPCClient) RaiseEventWorkflowBeta1(ctx context.Context, req *RaiseEventWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to raise event on workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to raise event on workflow: WorkflowComponent must be supplied")
	}
	if req.EventName == "" {
		return errors.New("failed to raise event on workflow: EventName must be supplied")
	}

	var eventData []byte
	var err error
	if (!req.SendRawData) && (req.EventData != nil) {
		eventData, err = json.Marshal(req.EventData)
		if err != nil {
			return fmt.Errorf("failed to marshal input: %v", err)
		}
	} else {
		eventData = []byte(fmt.Sprintf("%v", req.EventData))
	}

	_, err = c.protoClient.RaiseEventWorkflowBeta1(c.withAuthToken(ctx), &pb.RaiseEventWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
		EventName:         req.EventName,
		EventData:         eventData,
	})
	if err != nil {
		return fmt.Errorf("failed to raise event on workflow: %v", err)
	}
	return nil
}
