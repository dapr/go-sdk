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
	"fmt"
	"log"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
)

var stage = 0

const (
	workflowComponent = "dapr"
)

func main() {
	w, err := workflow.NewWorker()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Worker initialized")

	if err := w.RegisterWorkflow(TestWorkflow); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestWorkflow registered")

	if err := w.RegisterActivity(TestActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestActivity registered")

	// Start workflow runner
	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("runner started")

	daprClient, err := client.NewClient()
	if err != nil {
		log.Fatalf("failed to intialise client: %v", err)
	}
	defer daprClient.Close()
	ctx := context.Background()

	// Start workflow test
	respStart, err := daprClient.StartWorkflowBeta1(ctx, &client.StartWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
		WorkflowName:      "TestWorkflow",
		Options:           nil,
		Input:             1,
		SendRawInput:      false,
	})
	if err != nil {
		log.Fatalf("failed to start workflow: %v", err)
	}
	fmt.Printf("workflow started with id: %v\n", respStart.InstanceID)

	// Pause workflow test
	err = daprClient.PauseWorkflowBeta1(ctx, &client.PauseWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})

	if err != nil {
		log.Fatalf("failed to pause workflow: %v", err)
	}

	respGet, err := daprClient.GetWorkflowBeta1(ctx, &client.GetWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	if respGet.RuntimeStatus != workflow.StatusSuspended.String() {
		log.Fatalf("workflow not paused: %v", respGet.RuntimeStatus)
	}

	fmt.Printf("workflow paused\n")

	// Resume workflow test
	err = daprClient.ResumeWorkflowBeta1(ctx, &client.ResumeWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})

	if err != nil {
		log.Fatalf("failed to resume workflow: %v", err)
	}

	respGet, err = daprClient.GetWorkflowBeta1(ctx, &client.GetWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	if respGet.RuntimeStatus != workflow.StatusRunning.String() {
		log.Fatalf("workflow not running")
	}

	fmt.Println("workflow resumed")

	fmt.Printf("stage: %d\n", stage)

	// Raise Event Test

	err = daprClient.RaiseEventWorkflowBeta1(ctx, &client.RaiseEventWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
		EventName:         "testEvent",
		EventData:         "testData",
		SendRawData:       false,
	})

	if err != nil {
		fmt.Printf("failed to raise event: %v", err)
	}

	fmt.Println("workflow event raised")

	time.Sleep(time.Second) // allow workflow to advance

	fmt.Printf("stage: %d\n", stage)

	respGet, err = daprClient.GetWorkflowBeta1(ctx, &client.GetWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}

	fmt.Printf("workflow status: %v\n", respGet.RuntimeStatus)

	// Purge workflow test
	err = daprClient.PurgeWorkflowBeta1(ctx, &client.PurgeWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}

	respGet, err = daprClient.GetWorkflowBeta1(ctx, &client.GetWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil && respGet != nil {
		log.Fatal("failed to purge workflow")
	}

	fmt.Println("workflow purged")

	fmt.Printf("stage: %d\n", stage)

	// Terminate workflow test
	respStart, err = daprClient.StartWorkflowBeta1(ctx, &client.StartWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
		WorkflowName:      "TestWorkflow",
		Options:           nil,
		Input:             1,
		SendRawInput:      false,
	})
	if err != nil {
		log.Fatalf("failed to start workflow: %v", err)
	}

	fmt.Printf("workflow started with id: %s\n", respStart.InstanceID)

	err = daprClient.TerminateWorkflowBeta1(ctx, &client.TerminateWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil {
		log.Fatalf("failed to terminate workflow: %v", err)
	}

	respGet, err = daprClient.GetWorkflowBeta1(ctx, &client.GetWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}
	if respGet.RuntimeStatus != workflow.StatusTerminated.String() {
		log.Fatal("failed to terminate workflow")
	}

	fmt.Println("workflow terminated")

	err = daprClient.PurgeWorkflowBeta1(ctx, &client.PurgeWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})

	respGet, err = daprClient.GetWorkflowBeta1(ctx, &client.GetWorkflowRequest{
		InstanceID:        "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9",
		WorkflowComponent: workflowComponent,
	})
	if err == nil || respGet != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}

	fmt.Println("workflow purged")

	// WFClient
	// TODO: Expand client validation

	stage = 0
	fmt.Println("workflow client test")

	wfClient, err := workflow.NewClient()
	if err != nil {
		log.Fatalf("[wfclient] faield to initialize: %v", err)
	}

	id, err := wfClient.ScheduleNewWorkflow(ctx, "TestWorkflow", workflow.WithInstanceID("a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9"), workflow.WithInput(1))
	if err != nil {
		log.Fatalf("[wfclient] failed to start workflow: %v", err)
	}

	fmt.Printf("[wfclient] started workflow with id: %s\n", id)

	metadata, err := wfClient.FetchWorkflowMetadata(ctx, id)
	if err != nil {
		log.Fatalf("[wfclient] failed to get worfklow: %v", err)
	}

	fmt.Printf("[wfclient] workflow status: %v\n", metadata.RuntimeStatus.String())

	if stage != 1 {
		log.Fatalf("Workflow assertion failed while validating the wfclient. Stage 1 expected, current: %d", stage)
	}

	fmt.Printf("[wfclient] stage: %d\n", stage)

	// TODO: WaitForWorkflowStart
	// TODO: WaitForWorkflowCompletion

	// raise event

	if err := wfClient.RaiseEvent(ctx, id, "testEvent", workflow.WithEventPayload("testData")); err != nil {
		log.Fatalf("[wfclient] failed to raise event: %v", err)
	}

	fmt.Println("[wfclient] event raised")

	// Sleep to allow the workflow to advance
	time.Sleep(time.Second)

	if stage != 2 {
		log.Fatalf("Workflow assertion failed while validating the wfclient. Stage 2 expected, current: %d", stage)
	}

	fmt.Printf("[wfclient] stage: %d\n", stage)

	// stop workflow
	if err := wfClient.TerminateWorkflow(ctx, id); err != nil {
		log.Fatalf("[wfclient] failed to terminate workflow: %v", err)
	}

	fmt.Println("[wfclient] workflow terminated")

	if err := wfClient.PurgeWorkflow(ctx, id); err != nil {
		log.Fatalf("[wfclient] failed to purge workflow: %v", err)
	}

	fmt.Println("[wfclient] workflow purged")

	// stop workflow runtime
	if err := w.Shutdown(); err != nil {
		log.Fatalf("failed to shutdown runtime: %v", err)
	}

	fmt.Println("workflow worker successfully shutdown")
}

func TestWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var input int
	if err := ctx.GetInput(&input); err != nil {
		return nil, err
	}
	var output string
	if err := ctx.CallActivity(TestActivity, workflow.ActivityInput(input)).Await(&output); err != nil {
		return nil, err
	}

	err := ctx.WaitForExternalEvent("testEvent", time.Second*60).Await(&output)
	if err != nil {
		return nil, err
	}

	if err := ctx.CallActivity(TestActivity, workflow.ActivityInput(input)).Await(&output); err != nil {
		return nil, err
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
