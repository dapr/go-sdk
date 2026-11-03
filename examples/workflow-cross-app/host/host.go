package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

// The host app owns the workflow. It registers the workflow and its
// activities, starts the worker, and then simply stays alive. Every operation
// against the workflow instances it owns is driven by the separate client app,
// which never registers anything itself.
func main() {
	r := workflow.NewRegistry()

	if err := r.AddWorkflow(CrossAppWorkflow); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("host: CrossAppWorkflow registered")

	if err := r.AddActivity(GreetActivity); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("host: GreetActivity registered")

	if err := r.AddActivity(ConfirmActivity); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("host: ConfirmActivity registered")

	// The gRPC address of this app's own sidecar is picked up from the
	// DAPR_GRPC_PORT environment variable that `dapr run` injects.
	wfclient, err := client.NewWorkflowClient()
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := wfclient.StartWorker(ctx, r); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("host: worker started, waiting for cross-app operations")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	cancel()
	fmt.Println("host: worker shutdown")
}

// CrossAppWorkflow blocks on an external event so that the client app can
// meaningfully suspend, resume, raise an event against, and terminate it.
func CrossAppWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var name string
	if err := ctx.GetInput(&name); err != nil {
		return nil, err
	}

	var greeting string
	if err := ctx.CallActivity(GreetActivity, workflow.WithActivityInput(name)).Await(&greeting); err != nil {
		return nil, err
	}

	var approval string
	if err := ctx.WaitForExternalEvent("approval", 5*time.Minute).Await(&approval); err != nil {
		return nil, err
	}

	var confirmation string
	if err := ctx.CallActivity(ConfirmActivity, workflow.WithActivityInput(greeting+", "+approval)).Await(&confirmation); err != nil {
		return nil, err
	}

	return confirmation, nil
}

func GreetActivity(ctx workflow.ActivityContext) (any, error) {
	var name string
	if err := ctx.GetInput(&name); err != nil {
		return nil, err
	}

	return "hello " + name, nil
}

func ConfirmActivity(ctx workflow.ActivityContext) (any, error) {
	var input string
	if err := ctx.GetInput(&input); err != nil {
		return nil, err
	}

	return "confirmed: " + input, nil
}
