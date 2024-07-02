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
package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/microsoft/durabletask-go/api"
	"github.com/microsoft/durabletask-go/backend"
	durabletaskclient "github.com/microsoft/durabletask-go/client"

	dapr "github.com/dapr/go-sdk/client"
)

type client struct {
	taskHubClient *durabletaskclient.TaskHubGrpcClient
}

// WithInstanceID is an option to set an InstanceID when scheduling a new workflow.
func WithInstanceID(id string) api.NewOrchestrationOptions {
	return api.WithInstanceID(api.InstanceID(id))
}

// TODO: Implement WithOrchestrationIdReusePolicy

// WithInput is an option to pass an input when scheduling a new workflow.
func WithInput(input any) api.NewOrchestrationOptions {
	return api.WithInput(input)
}

// WithRawInput is an option to pass a byte slice as an input when scheduling a new workflow.
func WithRawInput(input string) api.NewOrchestrationOptions {
	return api.WithRawInput(input)
}

// WithStartTime is an option to set the start time when scheduling a new workflow.
func WithStartTime(time time.Time) api.NewOrchestrationOptions {
	return api.WithStartTime(time)
}

// WithFetchPayloads is an option to return the payload from a workflow.
func WithFetchPayloads(fetchPayloads bool) api.FetchOrchestrationMetadataOptions {
	return api.WithFetchPayloads(fetchPayloads)
}

// WithEventPayload is an option to send a payload with an event to a workflow.
func WithEventPayload(data any) api.RaiseEventOptions {
	return api.WithEventPayload(data)
}

// WithRawEventData is an option to send a byte slice with an event to a workflow.
func WithRawEventData(data string) api.RaiseEventOptions {
	return api.WithRawEventData(data)
}

// WithOutput is an option to define an output when terminating a workflow.
func WithOutput(data any) api.TerminateOptions {
	return api.WithOutput(data)
}

// WithRawOutput is an option to define a byte slice to output when terminating a workflow.
func WithRawOutput(data string) api.TerminateOptions {
	return api.WithRawOutput(data)
}

type clientOption func(*clientOptions) error

type clientOptions struct {
	daprClient dapr.Client
}

// WithDaprClient is an option to supply a custom dapr.Client to the workflow client.
func WithDaprClient(input dapr.Client) clientOption {
	return func(opt *clientOptions) error {
		opt.daprClient = input
		return nil
	}
}

// TODO: Implement mocks

// NewClient returns a workflow client.
func NewClient(opts ...clientOption) (client, error) {
	options := new(clientOptions)
	for _, configure := range opts {
		if err := configure(options); err != nil {
			return client{}, fmt.Errorf("failed to load options: %v", err)
		}
	}
	var daprClient dapr.Client
	var err error
	if options.daprClient == nil {
		daprClient, err = dapr.NewClient()
	} else {
		daprClient = options.daprClient
	}
	if err != nil {
		return client{}, fmt.Errorf("failed to initialise dapr.Client: %v", err)
	}

	taskHubClient := durabletaskclient.NewTaskHubGrpcClient(daprClient.GrpcClientConn(), backend.DefaultLogger())

	return client{
		taskHubClient: taskHubClient,
	}, nil
}

// ScheduleNewWorkflow will start a workflow and return the ID and/or error.
func (c *client) ScheduleNewWorkflow(ctx context.Context, workflow string, opts ...api.NewOrchestrationOptions) (id string, err error) {
	if workflow == "" {
		return "", errors.New("no workflow specified")
	}
	workflowID, err := c.taskHubClient.ScheduleNewOrchestration(ctx, workflow, opts...)
	return string(workflowID), err
}

// FetchWorkflowMetadata will return the metadata for a given workflow InstanceID and/or error.
func (c *client) FetchWorkflowMetadata(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*Metadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	wfMetadata, err := c.taskHubClient.FetchOrchestrationMetadata(ctx, api.InstanceID(id), opts...)
	if err != nil {
		return nil, err
	}

	return convertMetadata(wfMetadata), err
}

// WaitForWorkflowStart will wait for a given workflow to start and return metadata and/or an error.
func (c *client) WaitForWorkflowStart(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*Metadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	wfMetadata, err := c.taskHubClient.WaitForOrchestrationStart(ctx, api.InstanceID(id), opts...)
	if err != nil {
		return nil, err
	}

	return convertMetadata(wfMetadata), err
}

// WaitForWorkflowCompletion will block pending the completion of a specified workflow and return the metadata and/or error.
func (c *client) WaitForWorkflowCompletion(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*Metadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	wfMetadata, err := c.taskHubClient.WaitForOrchestrationCompletion(ctx, api.InstanceID(id), opts...)
	if err != nil {
		return nil, err
	}

	return convertMetadata(wfMetadata), err
}

// TerminateWorkflow will stop a given workflow and return an error output.
func (c *client) TerminateWorkflow(ctx context.Context, id string, opts ...api.TerminateOptions) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.TerminateOrchestration(ctx, api.InstanceID(id), opts...)
}

// RaiseEvent will raise an event on a given workflow and return an error output.
func (c *client) RaiseEvent(ctx context.Context, id, eventName string, opts ...api.RaiseEventOptions) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	if eventName == "" {
		return errors.New("no event name specified")
	}
	return c.taskHubClient.RaiseEvent(ctx, api.InstanceID(id), eventName, opts...)
}

// SuspendWorkflow will pause a given workflow and return an error output.
func (c *client) SuspendWorkflow(ctx context.Context, id, reason string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.SuspendOrchestration(ctx, api.InstanceID(id), reason)
}

// ResumeWorkflow will resume a suspended workflow and return an error output.
func (c *client) ResumeWorkflow(ctx context.Context, id, reason string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.ResumeOrchestration(ctx, api.InstanceID(id), reason)
}

// PurgeWorkflow will purge a given workflow and return an error output.
// NOTE: The workflow must be in a terminated or completed state.
func (c *client) PurgeWorkflow(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.PurgeOrchestrationState(ctx, api.InstanceID(id))
}
