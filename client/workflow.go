package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

type StartWorkflowRequest struct {
	InstanceID        string // Optional instance identifier
	WorkflowComponent string
	WorkflowName      string
	Options           map[string]string // Optional metadata
	Input             []byte            // Optional input
	// TODO: support data serialization
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
	EventData         []byte // Optional data
}

// StartWorkflowAlpha1 starts a workflow instance using the alpha1 spec.
func (c *GRPCClient) StartWorkflowAlpha1(ctx context.Context, req *StartWorkflowRequest) (*StartWorkflowResponse, error) {
	if req.InstanceID == "" {
		req.InstanceID = uuid.New().String()
	}
	if req.WorkflowComponent == "" {
		return nil, errors.New("failed to start workflow: WorkflowComponent must be supplied")
	}

	if req.WorkflowName == "" {
		return nil, errors.New("failed to start workflow: WorkflowName must be supplied")
	}
	resp, err := c.protoClient.StartWorkflowAlpha1(ctx, &pb.StartWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
		WorkflowName:      req.WorkflowName,
		Options:           req.Options,
		Input:             req.Input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start workflow instance: %v", err)
	}
	return &StartWorkflowResponse{
		InstanceID: resp.InstanceId,
	}, nil
}

// GetWorkflowAlpha1 gets the status of a workflow using the alpha1 spec.
func (c *GRPCClient) GetWorkflowAlpha1(ctx context.Context, req *GetWorkflowRequest) (*GetWorkflowResponse, error) {
	if req.InstanceID == "" {
		return nil, errors.New("failed to get workflow status: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return nil, errors.New("failed to get workflow status: WorkflowComponent must be supplied")
	}
	resp, err := c.protoClient.GetWorkflowAlpha1(ctx, &pb.GetWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %v", err)
	}
	return &GetWorkflowResponse{
		InstanceID:    resp.InstanceId,
		WorkflowName:  resp.WorkflowName,
		CreatedAt:     resp.CreatedAt.AsTime(),
		LastUpdatedAt: resp.LastUpdatedAt.AsTime(),
		RuntimeStatus: resp.RuntimeStatus,
		Properties:    resp.Properties,
	}, nil
}

// PurgeWorkflowAlpha1 removes all metadata relating to a specific workflow using the alpha1 spec.
func (c *GRPCClient) PurgeWorkflowAlpha1(ctx context.Context, req *PurgeWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to purge workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to purge workflow: WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.PurgeWorkflowAlpha1(ctx, &pb.PurgeWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to purge workflow: %v", err)
	}
	return nil
}

// TerminateWorkflowAlpha1 stops a workflow using the alpha1 spec.
func (c *GRPCClient) TerminateWorkflowAlpha1(ctx context.Context, req *TerminateWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to terminate workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to terminate workflow: WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.TerminateWorkflowAlpha1(ctx, &pb.TerminateWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to terminate workflow: %v", err)
	}
	return nil
}

// PauseWorkflowAlpha1 pauses a workflow that can be resumed later using the alpha1 spec.
func (c *GRPCClient) PauseWorkflowAlpha1(ctx context.Context, req *PauseWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to pause workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to pause workflow: WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.PauseWorkflowAlpha1(ctx, &pb.PauseWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to pause workflow: %v", err)
	}
	return nil
}

// ResumeWorkflowAlpha1 resumes a paused workflow using the alpha1 spec.
func (c *GRPCClient) ResumeWorkflowAlpha1(ctx context.Context, req *ResumeWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to resume workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to resume workflow: WorkflowComponent must be supplied")
	}
	_, err := c.protoClient.ResumeWorkflowAlpha1(ctx, &pb.ResumeWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return fmt.Errorf("failed to resume workflow: %v", err)
	}
	return nil
}

// RaiseEventWorkflowAlpha1 raises an event on a workflow using the alpha1 spec.
func (c *GRPCClient) RaiseEventWorkflowAlpha1(ctx context.Context, req *RaiseEventWorkflowRequest) error {
	if req.InstanceID == "" {
		return errors.New("failed to raise event on workflow: InstanceID must be supplied")
	}
	if req.WorkflowComponent == "" {
		return errors.New("failed to raise event on workflow: WorkflowComponent must be supplied")
	}
	if req.EventName == "" {
		return errors.New("failed to raise event on workflow: EventName must be supplied")
	}
	_, err := c.protoClient.RaiseEventWorkflowAlpha1(ctx, &pb.RaiseEventWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
		EventName:         req.EventName,
		EventData:         req.EventData,
	})
	if err != nil {
		return fmt.Errorf("failed to raise event on workflow: %v", err)
	}
	return nil
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
	resp, err := c.protoClient.StartWorkflowBeta1(ctx, &pb.StartWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
		WorkflowName:      req.WorkflowName,
		Options:           req.Options,
		Input:             req.Input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start workflow instance: %v", err)
	}
	return &StartWorkflowResponse{
		InstanceID: resp.InstanceId,
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
	resp, err := c.protoClient.GetWorkflowBeta1(ctx, &pb.GetWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %v", err)
	}
	return &GetWorkflowResponse{
		InstanceID:    resp.InstanceId,
		WorkflowName:  resp.WorkflowName,
		CreatedAt:     resp.CreatedAt.AsTime(),
		LastUpdatedAt: resp.LastUpdatedAt.AsTime(),
		RuntimeStatus: resp.RuntimeStatus,
		Properties:    resp.Properties,
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
	_, err := c.protoClient.PurgeWorkflowBeta1(ctx, &pb.PurgeWorkflowRequest{
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
	_, err := c.protoClient.ResumeWorkflowBeta1(ctx, &pb.ResumeWorkflowRequest{
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
	_, err := c.protoClient.RaiseEventWorkflowBeta1(ctx, &pb.RaiseEventWorkflowRequest{
		InstanceId:        req.InstanceID,
		WorkflowComponent: req.WorkflowComponent,
		EventName:         req.EventName,
		EventData:         req.EventData,
	})
	if err != nil {
		return fmt.Errorf("failed to raise event on workflow: %v", err)
	}
	return nil
}
