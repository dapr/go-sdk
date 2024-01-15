package workflow

import (
	"context"
	"errors"
	"time"

	"github.com/microsoft/durabletask-go/api"
	"github.com/microsoft/durabletask-go/backend"
	durabletaskclient "github.com/microsoft/durabletask-go/client"

	dapr "github.com/dapr/go-sdk/client"
)

type Client interface {
	ScheduleNewWorkflow(ctx context.Context) (string, error)
	FetchWorkflowMetadata(ctx context.Context) (string, error)
	WaitForWorkflowStart(ctx context.Context) (string, error)
	WaitForWorkflowCompletion(ctx context.Context) (string, error)
	TerminateWorkflow(ctx context.Context) error
	RaiseEvent(ctx context.Context) error
	SuspendWorkflow(ctx context.Context) error
	ResumeWorkflow(ctx context.Context) error
	PurgeWorkflow(ctx context.Context) error
}

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

// TODO: Implement mocks

func NewClient() (client, error) { // TODO: Implement custom connection
	daprClient, err := dapr.NewClient()
	if err != nil {
		return client{}, err
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
	if err != nil {
		return "", err
	}
	return string(workflowID), nil
}

func (c *client) FetchWorkflowMetadata(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*api.OrchestrationMetadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	return c.taskHubClient.FetchOrchestrationMetadata(ctx, api.InstanceID(id), opts...)
}

func (c *client) WaitForWorkflowStart(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*api.OrchestrationMetadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	return c.taskHubClient.WaitForOrchestrationStart(ctx, api.InstanceID(id), opts...)
}

func (c *client) WaitForWorkflowCompletion(ctx context.Context, id string, opts ...api.FetchOrchestrationMetadataOptions) (*api.OrchestrationMetadata, error) {
	if id == "" {
		return nil, errors.New("no workflow id specified")
	}
	return c.taskHubClient.WaitForOrchestrationCompletion(ctx, api.InstanceID(id), opts...)
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
	if reason == "" {
		return errors.New("no reason specified")
	}
	return c.taskHubClient.SuspendOrchestration(ctx, api.InstanceID(id), reason)
}

func (c *client) ResumeWorkflow(ctx context.Context, id, reason string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	if reason == "" {
		return errors.New("no reason specified")
	}
	return c.taskHubClient.ResumeOrchestration(ctx, api.InstanceID(id), reason)
}

func (c *client) PurgeWorkflow(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("no workflow id specified")
	}
	return c.taskHubClient.PurgeOrchestrationState(ctx, api.InstanceID(id))
}
