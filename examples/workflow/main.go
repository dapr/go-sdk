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
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

var stage = 0
var failActivityTries = 0

func main() {
	r := workflow.NewRegistry()

	if err := r.AddWorkflow(TestWorkflow); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestWorkflow registered")

	if err := r.AddActivity(TestActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestActivity registered")

	if err := r.AddActivity(FailActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("FailActivity registered")

	wclient, err := client.NewWorkflowClient()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Worker initialized")

	ctx, cancel := context.WithCancel(context.Background())
	if err = wclient.StartWorker(ctx, r); err != nil {
		log.Fatal(err)
	}
	fmt.Println("runner started")

	// Start workflow test
	// Set the start time to the current time to not wait for the workflow to
	// "start". This is useful for increasing the throughput of creating
	// workflows.
	// workflow.WithStartTime(time.Now())
	instanceID, err := wclient.ScheduleWorkflow(ctx, "TestWorkflow", workflow.WithInstanceID("a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9"), workflow.WithInput(1))
	if err != nil {
		log.Fatalf("failed to start workflow: %v", err)
	}
	fmt.Printf("workflow started with id: %v\n", instanceID)

	// Pause workflow test
	err = wclient.SuspendWorkflow(ctx, instanceID, "")
	if err != nil {
		log.Fatalf("failed to pause workflow: %v", err)
	}

	respFetch, err := wclient.FetchWorkflowMetadata(ctx, instanceID, workflow.WithFetchPayloads(true))
	if err != nil {
		log.Fatalf("failed to fetch workflow: %v", err)
	}

	if respFetch.RuntimeStatus != workflow.StatusSuspended {
		log.Fatalf("workflow not paused: %s: %v", respFetch.RuntimeStatus, respFetch)
	}

	fmt.Printf("workflow paused\n")

	// Resume workflow test
	err = wclient.ResumeWorkflow(ctx, instanceID, "")
	if err != nil {
		log.Fatalf("failed to resume workflow: %v", err)
	}

	respFetch, err = wclient.FetchWorkflowMetadata(ctx, instanceID, workflow.WithFetchPayloads(true))
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	if respFetch.RuntimeStatus != workflow.StatusRunning {
		log.Fatalf("workflow not running")
	}

	fmt.Println("workflow resumed")

	fmt.Printf("stage: %d\n", stage)

	// Raise Event Test

	err = wclient.RaiseEvent(ctx, instanceID, "testEvent", workflow.WithEventPayload("testData"))
	if err != nil {
		fmt.Printf("failed to raise event: %v", err)
	}

	fmt.Println("workflow event raised")

	time.Sleep(time.Second) // allow workflow to advance

	fmt.Printf("stage: %d\n", stage)

	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	_, err = wclient.WaitForWorkflowCompletion(waitCtx, instanceID)
	cancel()
	if err != nil {
		log.Fatalf("failed to wait for workflow: %v", err)
	}

	fmt.Printf("fail activity executions: %d\n", failActivityTries)

	respFetch, err = wclient.FetchWorkflowMetadata(ctx, instanceID, workflow.WithFetchPayloads(true))
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	fmt.Printf("workflow status: %v\n", respFetch.String())

	// Purge workflow test
	err = wclient.PurgeWorkflowState(ctx, instanceID)
	if err != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}

	respFetch, err = wclient.FetchWorkflowMetadata(ctx, instanceID, workflow.WithFetchPayloads(true))
	if err == nil || respFetch != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}

	fmt.Println("workflow purged")

	fmt.Printf("stage: %d\n", stage)

	// Terminate workflow test
	id, err := wclient.ScheduleWorkflow(ctx, "TestWorkflow", workflow.WithInstanceID("a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9"), workflow.WithInput(1))
	if err != nil {
		log.Fatalf("failed to start workflow: %v", err)
	}
	fmt.Printf("workflow started with id: %v\n", instanceID)

	metadata, err := wclient.WaitForWorkflowStart(ctx, id)
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

	cancel()

	fmt.Println("workflow worker successfully shutdown")
}

func TestWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var input int
	if err := ctx.GetInput(&input); err != nil {
		return nil, err
	}
	var output string
	if err := ctx.CallActivity(TestActivity, workflow.WithActivityInput(input)).Await(&output); err != nil {
		return nil, err
	}

	err := ctx.WaitForExternalEvent("testEvent", time.Second*60).Await(&output)
	if err != nil {
		return nil, err
	}

	if err := ctx.CallActivity(TestActivity, workflow.WithActivityInput(input)).Await(&output); err != nil {
		return nil, err
	}

	if err := ctx.CallActivity(FailActivity, workflow.WithActivityRetryPolicy(&workflow.RetryPolicy{
		MaxAttempts:          3,
		InitialRetryInterval: 100 * time.Millisecond,
		BackoffCoefficient:   2,
		MaxRetryInterval:     1 * time.Second,
	})).Await(nil); err == nil {
		return nil, fmt.Errorf("unexpected no error executing fail activity")
	}

	return output, nil
}

func TestActivity(ctx workflow.ActivityContext) (any, error) {
	var input int
	if err := ctx.GetInput(&input); err != nil {
		return "", err
	}

	stage += input

	return fmt.Sprintf("Stage: %d", stage), nil
}

func FailActivity(ctx workflow.ActivityContext) (any, error) {
	failActivityTries += 1
	return nil, errors.New("dummy activity error")
}
