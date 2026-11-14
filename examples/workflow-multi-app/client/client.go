package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dapr/durabletask-go/api"
	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

const (
	// hostAppID is the app ID of the app that registered MultiAppWorkflow and
	// runs the worker. Every client call below targets it explicitly.
	hostAppID = "multi-app-host"

	firstInstanceID  = "multi-app-workflow-1"
	secondInstanceID = "multi-app-workflow-2"
)

// The client app drives the whole workflow lifecycle on a different app. It
// has no registry and no worker of its own: the workflow name it schedules is
// only known to the host app.
func main() {
	// The gRPC address of this app's own sidecar is picked up from the
	// DAPR_GRPC_PORT environment variable that `dapr run` injects. The sidecar
	// then routes each operation on to the host app.
	wfclient, err := client.NewWorkflowClient()
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	runLifecycle(ctx, wfclient)
	runTerminate(ctx, wfclient)

	fmt.Println("client: all multi-app operations completed")
}

// runLifecycle schedules a workflow on the host app, then suspends, resumes,
// raises an event on, waits for and purges it, all by target app ID.
func runLifecycle(ctx context.Context, wfclient *workflow.Client) {
	id := schedule(ctx, wfclient, firstInstanceID)
	waitForStart(ctx, wfclient, id)

	if err := wfclient.SuspendWorkflow(ctx, id, "suspended from another app",
		workflow.WithSuspendAppID(hostAppID),
	); err != nil {
		logger.Fatalf("failed to suspend workflow: %v", err)
	}
	waitForStatus(ctx, wfclient, id, "SUSPENDED")
	fmt.Printf("client: workflow %s suspended on app %s\n", id, hostAppID)

	if err := wfclient.ResumeWorkflow(ctx, id, "resumed from another app",
		workflow.WithResumeAppID(hostAppID),
	); err != nil {
		logger.Fatalf("failed to resume workflow: %v", err)
	}
	waitForStatus(ctx, wfclient, id, "RUNNING")
	fmt.Printf("client: workflow %s resumed on app %s\n", id, hostAppID)

	// The workflow is parked on an external event, which only the host app can
	// deliver, so the event has to be routed there too.
	if err := wfclient.RaiseEvent(ctx, id, "approval",
		workflow.WithEventPayload("approved"),
		workflow.WithRaiseEventAppID(hostAppID),
	); err != nil {
		logger.Fatalf("failed to raise event: %v", err)
	}
	fmt.Printf("client: raised event approval on workflow %s on app %s\n", id, hostAppID)

	output := waitForCompletion(ctx, wfclient, id)
	fmt.Printf("client: workflow %s completed on app %s with output: %s\n", id, hostAppID, output)

	purge(ctx, wfclient, id)
}

// runTerminate schedules a second workflow on the host app and kills it while
// it is still parked on the external event.
func runTerminate(ctx context.Context, wfclient *workflow.Client) {
	id := schedule(ctx, wfclient, secondInstanceID)
	waitForStart(ctx, wfclient, id)

	if err := wfclient.TerminateWorkflow(ctx, id, workflow.WithTerminateAppID(hostAppID)); err != nil {
		logger.Fatalf("failed to terminate workflow: %v", err)
	}
	waitForStatus(ctx, wfclient, id, "TERMINATED")
	fmt.Printf("client: workflow %s terminated on app %s\n", id, hostAppID)

	purge(ctx, wfclient, id)
}

// schedule creates the instance on the host app rather than locally.
func schedule(ctx context.Context, wfclient *workflow.Client, instanceID string) string {
	id, err := wfclient.ScheduleWorkflow(ctx, "MultiAppWorkflow",
		workflow.WithAppID(hostAppID),
		workflow.WithInstanceID(instanceID),
		workflow.WithInput("dapr"),
	)
	if err != nil {
		logger.Fatalf("failed to schedule workflow: %v", err)
	}

	fmt.Printf("client: scheduled workflow %s on app %s\n", id, hostAppID)

	return id
}

// waitForStart blocks until the host app reports the instance as started.
// Fetching metadata for an instance owned by another app needs the fetch app
// ID, and so does every wait helper built on top of it.
func waitForStart(ctx context.Context, wfclient *workflow.Client, id string) {
	if _, err := wfclient.WaitForWorkflowStart(ctx, id, workflow.WithFetchAppID(hostAppID)); err != nil {
		logger.Fatalf("failed to wait for workflow start: %v", err)
	}

	fmt.Printf("client: workflow %s running on app %s\n", id, hostAppID)
}

// waitForCompletion blocks until the instance completes on the host app and
// returns its decoded output.
func waitForCompletion(ctx context.Context, wfclient *workflow.Client, id string) string {
	meta, err := wfclient.WaitForWorkflowCompletion(ctx, id,
		workflow.WithFetchAppID(hostAppID),
		workflow.WithFetchPayloads(true),
	)
	if err != nil {
		logger.Fatalf("failed to wait for workflow completion: %v", err)
	}
	if meta.String() != "COMPLETED" {
		logger.Fatalf("workflow %s did not complete: %s", id, meta.String())
	}

	var output string
	if uerr := json.Unmarshal([]byte(meta.Output.GetValue()), &output); uerr != nil {
		logger.Fatalf("failed to decode workflow output: %v", uerr)
	}

	return output
}

// purge removes the instance state from the host app and checks it is gone.
func purge(ctx context.Context, wfclient *workflow.Client, id string) {
	if err := wfclient.PurgeWorkflowState(ctx, id, workflow.WithPurgeAppID(hostAppID)); err != nil {
		logger.Fatalf("failed to purge workflow: %v", err)
	}

	if _, err := wfclient.FetchWorkflowMetadata(ctx, id, workflow.WithFetchAppID(hostAppID)); !errors.Is(err, api.ErrInstanceNotFound) {
		logger.Fatalf("workflow %s still exists after purge: %v", id, err)
	}

	fmt.Printf("client: workflow %s purged on app %s\n", id, hostAppID)
}

// waitForStatus polls the host app until the instance reports the wanted
// runtime status. Suspend, resume and terminate are all applied
// asynchronously, so polling keeps the printed output deterministic.
func waitForStatus(ctx context.Context, wfclient *workflow.Client, id, want string) {
	for range 120 {
		meta, err := wfclient.FetchWorkflowMetadata(ctx, id, workflow.WithFetchAppID(hostAppID))
		if err != nil {
			logger.Fatalf("failed to fetch workflow metadata: %v", err)
		}
		if meta.String() == want {
			return
		}

		select {
		case <-ctx.Done():
			logger.Fatalf("context done waiting for workflow %s to be %s", id, want)
		case <-time.After(250 * time.Millisecond):
		}
	}

	logger.Fatalf("workflow %s never reached status %s", id, want)
}
