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

// This example demonstrates the workflow schedule APIs:
//
//	CreateWorkflowSchedule  - start a workflow on a cron or interval schedule
//	GetWorkflowSchedule     - read a schedule and the status of its runs
//	ListWorkflowSchedules   - enumerate the app's schedules
//	TriggerWorkflowSchedule - start a run now, outside the schedule
//	DeleteWorkflowSchedule  - delete a schedule
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

const (
	scheduleName = "report-every-1s"
	manualName   = "report-manual"
)

func main() {
	r := workflow.NewRegistry()
	if err := r.AddWorkflow(ReportWorkflow); err != nil {
		logger.Fatal(err)
	}
	fmt.Println("ReportWorkflow registered")

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

	// Runs left by an earlier, interrupted execution would be listed and
	// counted with this one's.
	purgeRuns(ctx, wfClient, scheduleName, manualName)

	// Two runs one second apart, then the schedule completes. Overwrite makes
	// the example rerunnable: creating a schedule that already exists fails
	// with AlreadyExists otherwise.
	_, err = wfClient.CreateWorkflowSchedule(ctx, scheduleName, "ReportWorkflow", "@every 1s",
		workflow.WithScheduleInput("eu"),
		workflow.WithScheduleRepeats(2),
		workflow.WithScheduleOverwrite(),
	)
	if err != nil {
		logger.Fatalf("failed to create schedule %s: %v", scheduleName, err)
	}
	fmt.Printf("schedule created: %s\n", scheduleName)

	sched := waitForSchedule(ctx, wfClient, scheduleName, func(s *workflow.WorkflowSchedule) bool {
		return s.GetStatus().GetState() == workflow.WorkflowScheduleStateCompleted
	})
	fmt.Printf("schedule completed: runs=%d\n", sched.GetStatus().GetRunCount())

	// A schedule completes when its last run starts, not when it finishes.
	sched = waitForSchedule(ctx, wfClient, scheduleName, func(s *workflow.WorkflowSchedule) bool {
		return s.GetStatus().GetLastRun().GetRuntimeStatus() == workflow.StatusCompleted
	})
	fmt.Printf("last run status: %s\n",
		strings.TrimPrefix(sched.GetStatus().GetLastRun().GetRuntimeStatus().String(), "ORCHESTRATION_STATUS_"))

	// Each run is a workflow instance named "<schedule>_<owner>_<fire time>",
	// so the schedule name is a prefix to list them by.
	runs := runsOf(ctx, wfClient, scheduleName)
	if len(runs) == 0 {
		logger.Fatalf("no runs listed for %s", scheduleName)
	}
	var meta *workflow.WorkflowMetadata
	for _, id := range runs {
		if meta, err = wfClient.WaitForWorkflowCompletion(ctx, id, workflow.WithFetchPayloads(true)); err != nil {
			logger.Fatalf("failed to wait for run %s: %v", id, err)
		}
	}
	fmt.Printf("listed %d runs\n", len(runs))
	fmt.Printf("run output: %s\n", meta.Output.GetValue())

	// Like ListInstanceIDs, a page of schedules can be short while more remain,
	// so follow the continuation token to the end.
	listed := false
	var lopts []workflow.ListWorkflowSchedulesOptions
	for {
		list, lerr := wfClient.ListWorkflowSchedules(ctx, lopts...)
		if lerr != nil {
			logger.Fatalf("failed to list schedules: %v", lerr)
		}
		for _, s := range list.GetSchedules() {
			listed = listed || s.GetName() == scheduleName
		}
		if list.ContinuationToken == nil {
			break
		}
		lopts = []workflow.ListWorkflowSchedulesOptions{workflow.WithListWorkflowSchedulesContinuationToken(list.GetContinuationToken())}
	}
	fmt.Printf("schedule listed: %t\n", listed)

	// A trigger starts a run now and counts towards the schedule's run count,
	// without moving its next scheduled run.
	_, err = wfClient.CreateWorkflowSchedule(ctx, manualName, "ReportWorkflow", "@every 1h",
		workflow.WithScheduleInput("us"),
		workflow.WithScheduleOverwrite(),
	)
	if err != nil {
		logger.Fatalf("failed to create schedule %s: %v", manualName, err)
	}
	trig, err := wfClient.TriggerWorkflowSchedule(ctx, manualName)
	if err != nil {
		logger.Fatalf("failed to trigger schedule %s: %v", manualName, err)
	}
	if meta, err = wfClient.WaitForWorkflowCompletion(ctx, trig.GetInstanceId(), workflow.WithFetchPayloads(true)); err != nil {
		logger.Fatalf("failed to wait for triggered run %s: %v", trig.GetInstanceId(), err)
	}
	fmt.Printf("triggered run completed: %s\n", meta.Output.GetValue())

	got, err := wfClient.GetWorkflowSchedule(ctx, manualName)
	if err != nil {
		logger.Fatalf("failed to get schedule %s: %v", manualName, err)
	}
	fmt.Printf("manual schedule run count: %d\n", got.GetSchedule().GetStatus().GetRunCount())

	// Deleting a schedule leaves the runs it already started untouched.
	for _, name := range []string{scheduleName, manualName} {
		if _, err = wfClient.DeleteWorkflowSchedule(ctx, name); err != nil {
			logger.Fatalf("failed to delete schedule %s: %v", name, err)
		}
		if _, err = wfClient.GetWorkflowSchedule(ctx, name); status.Code(err) != codes.NotFound {
			logger.Fatalf("expected NotFound for deleted schedule %s, got: %v", name, err)
		}
	}
	fmt.Println("schedules deleted")

	purgeRuns(ctx, wfClient, scheduleName, manualName)
	fmt.Println("runs purged")

	fmt.Println("workflow schedule example completed")
}

// runsOf lists the instance IDs of a schedule's runs, which start with the
// schedule name and "_". daprd matches the prefix only up to its first LIKE
// wildcard, here the "_", and filters the rest, so a page can come back short,
// even empty, while more remain: follow the continuation token to the end.
func runsOf(ctx context.Context, wfClient *workflow.Client, schedule string) []string {
	prefix := workflow.WithListInstanceIDsInstanceIDPrefix(schedule + "_")
	opts := []workflow.ListInstanceIDsOptions{prefix}
	var ids []string
	for {
		resp, err := wfClient.ListInstanceIDs(ctx, opts...)
		if err != nil {
			logger.Fatalf("failed to list runs of %s: %v", schedule, err)
		}
		ids = append(ids, resp.InstanceIds...)
		// ListInstanceIDsResponse is a defined type over the protobuf message,
		// so it has no generated getters.
		if resp.ContinuationToken == nil {
			return ids
		}
		opts = []workflow.ListInstanceIDsOptions{prefix, workflow.WithListInstanceIDsContinuationToken(*resp.ContinuationToken)}
	}
}

// purgeRuns waits for every run of the schedules to finish, then purges it.
func purgeRuns(ctx context.Context, wfClient *workflow.Client, schedules ...string) {
	for _, schedule := range schedules {
		for _, id := range runsOf(ctx, wfClient, schedule) {
			if _, err := wfClient.WaitForWorkflowCompletion(ctx, id); err != nil {
				logger.Fatalf("failed to wait for run %s: %v", id, err)
			}
			if err := wfClient.PurgeWorkflowState(ctx, id); err != nil {
				logger.Fatalf("failed to purge run %s: %v", id, err)
			}
		}
	}
}

// waitForSchedule polls the schedule until ready reports it is, or 30s pass.
func waitForSchedule(ctx context.Context, wfClient *workflow.Client, name string, ready func(*workflow.WorkflowSchedule) bool) *workflow.WorkflowSchedule {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		resp, err := wfClient.GetWorkflowSchedule(ctx, name)
		if err != nil {
			logger.Fatalf("failed to get schedule %s: %v", name, err)
		}
		if ready(resp.GetSchedule()) {
			return resp.GetSchedule()
		}

		select {
		case <-ctx.Done():
			logger.Fatalf("timed out waiting for schedule %s, last status: %v", name, resp.GetSchedule().GetStatus())
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func ReportWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var region string
	if err := ctx.GetInput(&region); err != nil {
		return nil, err
	}

	return "report for " + region, nil
}
