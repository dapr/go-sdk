package main

import (
	"context"
	"log"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

func main() {
	ctx := context.Background()

	registry := workflow.NewTaskRegistry()
	if err := registry.AddWorkflow(TestWorkflow); err != nil {
		log.Fatal(err)
	}

	daprClient, err := client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	wfclient := workflow.NewClient(daprClient.GrpcClientConn())
	if err := wfclient.StartWorker(ctx, registry); err != nil {
		log.Fatal(err)
	}

	id, err := wfclient.ScheduleNewWorkflow(ctx, "TestWorkflow")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = wfclient.WaitForWorkflowCompletion(ctx, id); err != nil {
		log.Fatal(err)
	}
}

func TestWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var output string
	err := ctx.CallChildWorkflow("my-sub-orchestration",
		workflow.WithChildWorkflowInput("my-input"),
		// Here we set custom target app ID which will execute this activity.
		workflow.WithChildWorkflowAppID("my-sub-app-id"),
	).Await(&output)
	if err != nil {
		return nil, err
	}

	return output, nil
}
