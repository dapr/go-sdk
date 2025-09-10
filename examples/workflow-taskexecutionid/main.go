package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

func main() {
	r := workflow.NewRegistry()

	if err := r.AddWorkflow(TaskExecutionIdWorkflow); err != nil {
		log.Fatalf("failed to register workflow: %v", err)
	}
	if err := r.AddActivity(RetryN); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}
	fmt.Println("Workflow(s) and activities registered.")

	wclient, err := client.NewWorkflowClient()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Worker initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = wclient.StartWorker(ctx, r); err != nil {
		log.Fatal(err)
	}

	id, err := wclient.ScheduleWorkflow(ctx, "TaskExecutionIdWorkflow", workflow.WithInput(5))
	if err != nil {
		log.Fatalf("failed to schedule a new workflow: %v", err)
	}

	metadata, err := wclient.WaitForWorkflowCompletion(ctx, id)
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}
	fmt.Printf("workflow status: %s\n", metadata.String())

	err = wclient.TerminateWorkflow(ctx, id)
	if err != nil {
		log.Fatalf("failed to terminate workflow: %v", err)
	}
	fmt.Println("workflow terminated")

	err = wclient.PurgeWorkflowState(ctx, id)
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
