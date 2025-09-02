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

	"github.com/dapr/durabletask-go/api"
	"github.com/dapr/durabletask-go/backend"
	"github.com/dapr/durabletask-go/client"
	"github.com/dapr/durabletask-go/task"
	dapr "github.com/dapr/go-sdk/client"
)

var stage = 0
var failActivityTries = 0

func main() {
	registry := task.NewTaskRegistry()

	if err := registry.AddOrchestrator(TestWorkflow); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestWorkflow registered")

	if err := registry.AddActivity(TestActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestActivity registered")

	if err := registry.AddActivity(FailActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("FailActivity registered")

	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("failed to create Dapr client: %v", err)
	}

	client := client.NewTaskHubGrpcClient(daprClient.GrpcClientConn(), backend.DefaultLogger())

	fmt.Println("Worker initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start workflow runner
	if err := client.StartWorkItemListener(ctx, registry); err != nil {
		log.Fatalf("failed to start work item listener: %v", err)
	}
	fmt.Println("runner started")

	// Start workflow test
	// Set the start time to the current time to not wait for the workflow to
	// "start". This is useful for increasing the throughput of creating
	// workflows.
	// workflow.WithStartTime(time.Now())
	instanceID, err := client.ScheduleNewOrchestration(ctx, "TestWorkflow",
		api.WithInstanceID("a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9"),
		api.WithInput(1),
	)
	if err != nil {
		log.Fatalf("failed to start workflow: %v", err)
	}
	fmt.Printf("workflow started with id: %v\n", instanceID)

	// Pause workflow test
	err = client.SuspendOrchestration(ctx, instanceID, "")
	if err != nil {
		log.Fatalf("failed to pause workflow: %v", err)
	}

	respFetch, err := client.FetchOrchestrationMetadata(ctx, instanceID, api.WithFetchPayloads(true))
	if err != nil {
		log.Fatalf("failed to fetch workflow: %v", err)
	}

	if respFetch.RuntimeStatus != api.RUNTIME_STATUS_SUSPENDED {
		log.Fatalf("workflow not paused: %v", respFetch.RuntimeStatus)
	}

	fmt.Printf("workflow paused\n")

	// Resume workflow test
	err = client.ResumeOrchestration(ctx, instanceID, "")
	if err != nil {
		log.Fatalf("failed to resume workflow: %v", err)
	}

	respFetch, err = client.FetchOrchestrationMetadata(ctx, instanceID, api.WithFetchPayloads(true))
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	if respFetch.RuntimeStatus != api.RUNTIME_STATUS_RUNNING {
		log.Fatalf("workflow not running")
	}

	fmt.Println("workflow resumed")

	fmt.Printf("stage: %d\n", stage)

	// Raise Event Test

	err = client.RaiseEvent(ctx, instanceID, "testEvent", api.WithEventPayload("testData"))
	if err != nil {
		fmt.Printf("failed to raise event: %v", err)
	}

	fmt.Println("workflow event raised")

	time.Sleep(time.Second) // allow workflow to advance

	fmt.Printf("stage: %d\n", stage)

	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	_, err = client.WaitForOrchestrationCompletion(waitCtx, instanceID)
	cancel()
	if err != nil {
		log.Fatalf("failed to wait for workflow: %v", err)
	}

	fmt.Printf("fail activity executions: %d\n", failActivityTries)

	respFetch, err = client.FetchOrchestrationMetadata(ctx, instanceID, api.WithFetchPayloads(true))
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	fmt.Printf("workflow status: %v\n", respFetch.RuntimeStatus)

	// Purge workflow test
	err = client.PurgeOrchestrationState(ctx, instanceID)
	if err != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}

	respFetch, err = client.FetchOrchestrationMetadata(ctx, instanceID, api.WithFetchPayloads(true))
	if err == nil || respFetch != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}

	fmt.Println("workflow purged")

	fmt.Printf("stage: %d\n", stage)

	// Terminate workflow test
	id, err := client.ScheduleNewOrchestration(ctx, "TestWorkflow", api.WithInstanceID("a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9"), api.WithInput(1))
	if err != nil {
		log.Fatalf("failed to start workflow: %v", err)
	}
	fmt.Printf("workflow started with id: %v\n", instanceID)

	metadata, err := client.WaitForOrchestrationStart(ctx, id)
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}
	fmt.Printf("workflow status: %s\n", metadata.RuntimeStatus.String())

	err = client.TerminateOrchestration(ctx, id)
	if err != nil {
		log.Fatalf("failed to terminate workflow: %v", err)
	}
	fmt.Println("workflow terminated")

	err = client.PurgeOrchestrationState(ctx, id)
	if err != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}
	fmt.Println("workflow purged")

	// stop workflow runtime
	cancel()

	fmt.Println("workflow worker successfully shutdown")
}

func TestWorkflow(ctx *task.OrchestrationContext) (any, error) {
	var input int
	if err := ctx.GetInput(&input); err != nil {
		return nil, err
	}
	var output string
	if err := ctx.CallActivity(TestActivity, task.WithActivityInput(input)).Await(&output); err != nil {
		return nil, err
	}

	err := ctx.WaitForSingleEvent("testEvent", time.Second*60).Await(&output)
	if err != nil {
		return nil, err
	}

	if err := ctx.CallActivity(TestActivity, task.WithActivityInput(input)).Await(&output); err != nil {
		return nil, err
	}

	if err := ctx.CallActivity(FailActivity, task.WithActivityRetryPolicy(&task.RetryPolicy{
		MaxAttempts:          3,
		InitialRetryInterval: 100 * time.Millisecond,
		BackoffCoefficient:   2,
		MaxRetryInterval:     1 * time.Second,
	})).Await(nil); err == nil {
		return nil, fmt.Errorf("unexpected no error executing fail activity")
	}

	return output, nil
}

func TestActivity(ctx task.ActivityContext) (any, error) {
	var input int
	if err := ctx.GetInput(&input); err != nil {
		return "", err
	}

	stage += input

	return fmt.Sprintf("Stage: %d", stage), nil
}

func FailActivity(ctx task.ActivityContext) (any, error) {
	failActivityTries += 1
	return nil, errors.New("dummy activity error")
}
