package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"log"
	"sync"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
)

var stage = 0

const (
	workflowComponent = "dapr"
)

func main() {
	wr, err := workflow.NewRuntime("localhost", "50001")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Runtime initialized")

	if err := wr.RegisterWorkflow(TestWorkflow); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestWorkflow registered")

	if err := wr.RegisterActivity(TestActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestActivity registered")

	var wg sync.WaitGroup

	// start workflow runner
	fmt.Println("runner 1")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := wr.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	daprClient, err := client.NewClient()
	defer daprClient.Close()
	if err != nil {
		log.Fatalf("failed to intialise client: %v", err)
	}
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "dapr-app-id", "workflow-sequential")

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

	fmt.Printf("workflow resumed\n")

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

	// stop workflow runtime
	if err := wr.Shutdown(); err != nil {
		log.Fatalf("failed to shutdown runtime: %v", err)
	}

	time.Sleep(time.Second * 5)
}

func TestWorkflow(ctx *workflow.Context) (any, error) {
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
