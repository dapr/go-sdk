/*
Copyright 2026 The Dapr Authors
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

// This example demonstrates the advanced workflow management APIs:
//
//	ListInstanceIDs        - enumerate instance IDs, with pagination
//	GetInstanceHistory     - read an instance's full execution history
//	RerunWorkflowFromEvent - rerun ("rewind") an instance from a history event
//
// The three compose naturally: list the instances, read one instance's history
// to find an event worth rerunning from, then rerun from it.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

// ListInstanceIDs returns every instance this app has in its state store, so
// the example filters on a known prefix. The example purges what it created on
// the way out, so consecutive successful runs see the same three instances; a
// run that fails part way through skips the purge and leaves state behind.
const idPrefix = "order-"

func main() {
	r := workflow.NewRegistry()

	if err := r.AddWorkflow(OrderWorkflow); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddActivity(ReserveStock); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddActivity(ChargeCard); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("OrderWorkflow registered")

	wfClient, err := client.NewWorkflowClient()
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = wfClient.StartWorker(ctx, r); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("workflow worker started")

	// Schedule a few instances so there is something to list and rerun.
	ids := []string{idPrefix + "01", idPrefix + "02", idPrefix + "03"}
	for _, id := range ids {
		if _, err = wfClient.ScheduleWorkflow(ctx, "OrderWorkflow",
			workflow.WithInstanceID(id),
			workflow.WithInput(10),
		); err != nil {
			logger.Fatalf("failed to schedule %s: %v", id, err)
		}
		if _, err = wfClient.WaitForWorkflowCompletion(ctx, id); err != nil {
			logger.Fatalf("failed to wait for %s: %v", id, err)
		}
	}
	fmt.Printf("scheduled and completed %d workflows\n", len(ids))

	instances := listInstances(ctx, wfClient)
	eventID, activity := findRerunTarget(ctx, wfClient, ids[0])
	rerunID := rerun(ctx, wfClient, ids[0], eventID, activity)

	// Purge everything this example created, including the rerun instance. A
	// leftover rerun instance from an earlier failed run also matches idPrefix,
	// so it can already be in `instances`; purging the same ID twice fails.
	for _, id := range dedupe(append(instances, rerunID)) {
		if err = wfClient.PurgeWorkflowState(ctx, id); err != nil {
			logger.Fatalf("failed to purge %s: %v", id, err)
		}
	}
	fmt.Println("purged all instances")

	// GetInstanceHistory reports NotFound once the state is gone.
	_, err = wfClient.GetInstanceHistory(ctx, ids[0])
	if status.Code(err) != codes.NotFound {
		logger.Fatalf("expected NotFound for purged instance %s, got: %v", ids[0], err)
	}
	fmt.Println("history for purged instance is gone")

	fmt.Println("workflow management example completed")
}

// listInstances pages through every instance ID the app has in its state store.
// The page size is an upper bound on a page: keep calling with the continuation
// token from the previous response until it comes back nil.
func listInstances(ctx context.Context, wfClient *workflow.Client) []string {
	fmt.Println("== ListInstanceIDs ==")

	var (
		found []string
		token string
		pages int
	)

	for {
		opts := []workflow.ListInstanceIDsOptions{
			workflow.WithListInstanceIDsPageSize(2),
		}
		if token != "" {
			opts = append(opts, workflow.WithListInstanceIDsContinuationToken(token))
		}

		resp, err := wfClient.ListInstanceIDs(ctx, opts...)
		if err != nil {
			logger.Fatalf("failed to list instance IDs: %v", err)
		}
		pages++

		for _, id := range resp.InstanceIds {
			if strings.HasPrefix(id, idPrefix) {
				found = append(found, id)
			}
		}

		// ListInstanceIDsResponse is a defined type over the generated protobuf
		// message, which means it does not inherit the generated getters. Read
		// the optional field as a pointer instead of GetContinuationToken().
		if resp.ContinuationToken == nil {
			break
		}
		token = *resp.ContinuationToken
	}

	sort.Strings(found)
	fmt.Printf("listed %d instances with prefix %s over %d page(s)\n", len(found), idPrefix, pages)
	for _, id := range found {
		fmt.Printf("  instance: %s\n", id)
	}

	return found
}

// findRerunTarget reads an instance's history and picks an event to rerun from.
//
// Only three event types can be rerun: a scheduled activity (TaskScheduled), a
// created timer (TimerCreated) and a created child workflow
// (ChildWorkflowInstanceCreated). Any other event is rejected with NotFound.
// The rerun targets an event's own EventId, not its index in the history.
func findRerunTarget(ctx context.Context, wfClient *workflow.Client, id string) (uint32, string) {
	fmt.Println("== GetInstanceHistory ==")

	hist, err := wfClient.GetInstanceHistory(ctx, id)
	if err != nil {
		logger.Fatalf("failed to get history for %s: %v", id, err)
	}
	fmt.Printf("instance %s has %d history events\n", id, len(hist.Events))

	var (
		eventID  uint32
		activity string
	)
	for _, e := range hist.Events {
		ts := e.GetTaskScheduled()
		if ts == nil {
			continue
		}
		fmt.Printf("  rerunnable event: type=TaskScheduled name=%s\n", ts.GetName())

		// History event IDs are int32 and some event types carry -1, while the
		// rerun API takes a uint32, so the ID needs narrowing before use.
		// Scheduled tasks are always numbered from zero.
		rawID := e.GetEventId()
		if rawID >= 0 {
			// Target the last activity. This workflow is sequential, so the
			// earlier activity completed before the target and is replayed from
			// history rather than executed again.
			eventID, activity = uint32(rawID), ts.GetName() //nolint:gosec // guarded above
		}
	}

	if activity == "" {
		logger.Fatalf("no rerunnable event found in the history of %s", id)
	}

	return eventID, activity
}

// rerun reruns an instance from a history event with a fresh input.
//
// The source instance must already be in a terminal state, otherwise the
// runtime rejects the request with InvalidArgument. WithRerunInput replaces the
// input of the target event only. Work completed and recorded before the target
// event is replayed from history; anything from before it that had not completed
// is re-executed in the new instance.
func rerun(ctx context.Context, wfClient *workflow.Client, id string, eventID uint32, activity string) string {
	fmt.Println("== RerunWorkflowFromEvent ==")

	rerunID, err := wfClient.RerunWorkflowFromEvent(ctx, id, eventID,
		workflow.WithRerunNewInstanceID(id+"-rerun"),
		workflow.WithRerunInput(25),
	)
	if err != nil {
		logger.Fatalf("failed to rerun %s from event %d: %v", id, eventID, err)
	}
	fmt.Printf("rerunning %s from its %s activity as %s\n", id, activity, rerunID)

	meta, err := wfClient.WaitForWorkflowCompletion(ctx, rerunID)
	if err != nil {
		logger.Fatalf("failed to wait for rerun %s: %v", rerunID, err)
	}
	// Printing the output is what pins WithRerunInput: ChargeCard reran with the
	// replacement input, so the workflow returns "charged 25" and not the
	// "charged 10" the source instance produced.
	fmt.Printf("rerun instance %s finished with status: %s, output: %s\n",
		rerunID, meta.String(), meta.Output.GetValue())

	return rerunID
}

// dedupe returns ids with duplicates removed, preserving order.
func dedupe(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}

	return out
}

func OrderWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var qty int
	if err := ctx.GetInput(&qty); err != nil {
		return nil, err
	}

	var reserved string
	if err := ctx.CallActivity(ReserveStock, workflow.WithActivityInput(qty)).Await(&reserved); err != nil {
		return nil, err
	}

	var charged string
	if err := ctx.CallActivity(ChargeCard, workflow.WithActivityInput(qty)).Await(&charged); err != nil {
		return nil, err
	}

	return charged, nil
}

func ReserveStock(ctx workflow.ActivityContext) (any, error) {
	var qty int
	if err := ctx.GetInput(&qty); err != nil {
		return nil, err
	}

	return fmt.Sprintf("reserved %d", qty), nil
}

func ChargeCard(ctx workflow.ActivityContext) (any, error) {
	var qty int
	if err := ctx.GetInput(&qty); err != nil {
		return nil, err
	}

	return fmt.Sprintf("charged %d", qty), nil
}
