package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dapr/go-sdk/workflow"
)

func main() {
	w, err := workflow.NewWorker()
	if err != nil {
		log.Fatalf("failed to initialise worker: %v", err)
	}

	if err := w.RegisterWorkflow(TaskExecutionIdWorkflow); err != nil {
		log.Fatalf("failed to register workflow: %v", err)
	}
	if err := w.RegisterActivity(RetryN); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}
	fmt.Println("Workflow(s) and activities registered.")

	if err := w.Start(); err != nil {
		log.Fatalf("failed to start worker")
	}

	wfClient, err := workflow.NewClient()
	if err != nil {
		log.Fatalf("failed to initialise client: %v", err)
	}
	ctx := context.Background()
	id, err := wfClient.ScheduleNewWorkflow(ctx, "TaskExecutionIdWorkflow", workflow.WithInput(5))
	if err != nil {
		log.Fatalf("failed to schedule a new workflow: %v", err)
	}

	metadata, err := wfClient.WaitForWorkflowCompletion(ctx, id)
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}
	fmt.Printf("workflow status: %s\n", metadata.RuntimeStatus.String())

	err = wfClient.TerminateWorkflow(ctx, id)
	if err != nil {
		log.Fatalf("failed to terminate workflow: %v", err)
	}
	fmt.Println("workflow terminated")

	err = wfClient.PurgeWorkflow(ctx, id)
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
	if err := ctx.CallActivity(RetryN, workflow.ActivityRetryPolicy(workflow.RetryPolicy{
		MaxAttempts:          retries,
		InitialRetryInterval: 100 * time.Millisecond,
		BackoffCoefficient:   2,
		MaxRetryInterval:     1 * time.Second,
	}), workflow.ActivityInput(retries)).Await(&workBatch); err != nil {
		return 0, err
	}

	if err := ctx.CallActivity(RetryN, workflow.ActivityRetryPolicy(workflow.RetryPolicy{
		MaxAttempts:          retries,
		InitialRetryInterval: 100 * time.Millisecond,
		BackoffCoefficient:   2,
		MaxRetryInterval:     1 * time.Second,
	}), workflow.ActivityInput(retries)).Await(&workBatch); err != nil {
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
