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

func WithInstanceID(id string) api.NewOrchestrationOptions {
	return api.WithInstanceID(api.InstanceID(id))
}

// TODO: Implement WithOrchestrationIdReusePolicy

func WithInput(input any) api.NewOrchestrationOptions {
	return api.WithInput(input)
}

func WithRawInput(input string) api.NewOrchestrationOptions {
	return api.WithRawInput(input)
}

func WithStartTime(time time.Time) api.NewOrchestrationOptions {
	return api.WithStartTime(time)
}

func WithFetchPayloads(fetchPayloads bool) api.FetchOrchestrationMetadataOptions {
	return api.WithFetchPayloads(fetchPayloads)
}

func WithEventPayload(data any) api.RaiseEventOptions {
	return api.WithEventPayload(data)
}

func WithRawEventData(data string) api.RaiseEventOptions {
	return api.WithRawEventData(data)
}

func WithOutput(data any) api.TerminateOptions {
	return api.WithOutput(data)
}

func WithRawOutput(data string) api.TerminateOptions {
	return api.WithRawOutput(data)
}

type clientOption func(*clientOptions) error

type clientOptions struct {
	daprClient dapr.Client
}

func WithDaprClient(input dapr.Client) clientOption {
	return func(opt *clientOptions) error {
		opt.daprClient = input
		return nil
	}
}

// TODO: Implement mocks

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

func (c *client) ScheduleNewWorkflow(ctx context.Context, workflow string, opts ...api.NewOrchestrationOptions) (id string, err error) {
	if workflow == "" {
		return "", errors.New("no workflow specified")
	}
	workflowID, err := c.taskHubClient.ScheduleNewOrchestration(ctx, workflow, opts...)
	return string(workflowID), err
}

func (c *client) FetchWorkflowMetadata(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*Metadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	wfMetadata, err := c.taskHubClient.FetchOrchestrationMetadata(ctx, api.InstanceID(id), opts...)

	return convertMetadata(wfMetadata), err
}

func (c *client) WaitForWorkflowStart(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*Metadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	wfMetadata, err := c.taskHubClient.WaitForOrchestrationStart(ctx, api.InstanceID(id), opts...)

	return convertMetadata(wfMetadata), err
}

func (c *client) WaitForWorkflowCompletion(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*Metadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	wfMetadata, err := c.taskHubClient.WaitForOrchestrationCompletion(ctx, api.InstanceID(id), opts...)

	return convertMetadata(wfMetadata), err
}

func (c *client) TerminateWorkflow(ctx context.Context, id string, opts ...api.TerminateOptions) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.TerminateOrchestration(ctx, api.InstanceID(id), opts...)
}

func (c *client) RaiseEvent(ctx context.Context, id, eventName string, opts ...api.RaiseEventOptions) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	if eventName == "" {
		return errors.New("no event name specified")
	}
	return c.taskHubClient.RaiseEvent(ctx, api.InstanceID(id), eventName, opts...)
}

func (c *client) SuspendWorkflow(ctx context.Context, id, reason string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.SuspendOrchestration(ctx, api.InstanceID(id), reason)
}

func (c *client) ResumeWorkflow(ctx context.Context, id, reason string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.ResumeOrchestration(ctx, api.InstanceID(id), reason)
}

func (c *client) PurgeWorkflow(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.PurgeOrchestrationState(ctx, api.InstanceID(id))
}
