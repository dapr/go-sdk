package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	registry := workflow.NewTaskRegistry()

	if err := registry.AddWorkflow(TaskExecutionIdWorkflow); err != nil {
		log.Fatalf("failed to register workflow: %v", err)
	}
	if err := registry.AddActivity(RetryN); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}
	fmt.Println("Workflow(s) and activities registered.")

	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("failed to create Dapr client: %v", err)
	}

	ctx := context.Background()
	wfclient := workflow.NewClient(daprClient.GrpcClientConn())
	if err := wfclient.StartWorker(ctx, registry); err != nil {
		log.Fatalf("failed to start work item listener: %v", err)
	}

	id, err := wfclient.ScheduleNewWorkflow(ctx, "TaskExecutionIdWorkflow", workflow.WithInput(5))
	if err != nil {
		log.Fatalf("failed to schedule a new workflow: %v", err)
	}

	metadata, err := wfclient.WaitForWorkflowCompletion(ctx, id)
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}
	fmt.Printf("workflow status: %s\n", metadata.RuntimeStatus.String())

	err = wfclient.TerminateWorkflow(ctx, id)
	if err != nil {
		log.Fatalf("failed to terminate workflow: %v", err)
	}
	fmt.Println("workflow terminated")

	err = wfclient.PurgeWorkflowState(ctx, id)
	if err != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}
	fmt.Println("workflow purged")
}

var eMap = sync.Map{}

func TaskExecutionIdWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var retries int
	if err := ctx.GetInput(&retries); err != nil {
		return 0, err
	}

	var workBatch []int
	if err := ctx.CallActivity(RetryN, workflow.WithActivityRetryPolicy(&workflow.RetryPolicy{
		MaxAttempts:          retries,
		InitialRetryInterval: 100 * time.Millisecond,
		BackoffCoefficient:   2,
		MaxRetryInterval:     1 * time.Second,
	}), workflow.WithActivityInput(retries)).Await(&workBatch); err != nil {
		return 0, err
	}

	if err := ctx.CallActivity(RetryN, workflow.WithActivityRetryPolicy(&workflow.RetryPolicy{
		MaxAttempts:          retries,
		InitialRetryInterval: 100 * time.Millisecond,
		BackoffCoefficient:   2,
		MaxRetryInterval:     1 * time.Second,
	}), workflow.WithActivityInput(retries)).Await(&workBatch); err != nil {
		return 0, err
	}

	return 0, nil
}

func RetryN(ctx workflow.ActivityContext) (any, error) {
	taskExecutionID := ctx.GetTaskExecutionID()
	counter, _ := eMap.LoadOrStore(taskExecutionID, &atomic.Int32{})
	var retries int32
	if err := ctx.GetInput(&retries); err != nil {
		return 0, err
	}

	counter.(*atomic.Int32).Add(1)
	fmt.Println("RetryN ", counter.(*atomic.Int32).Load())

	if counter.(*atomic.Int32).Load() < retries-1 {
		return nil, fmt.Errorf("failed")
	}

	return nil, nil

}
