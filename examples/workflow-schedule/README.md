# Dapr Workflow Schedule Example with go-sdk

Demonstrates the workflow schedule APIs exposed by the Dapr Go SDK:

| API | What it does |
| --- | --- |
| `CreateWorkflowSchedule` | Start a workflow on a cron or interval schedule |
| `GetWorkflowSchedule` | Read a schedule and the status of its runs |
| `ListWorkflowSchedules` | Enumerate the app's schedules, with pagination |
| `TriggerWorkflowSchedule` | Start a run now, outside the schedule |
| `DeleteWorkflowSchedule` | Delete a schedule |

The example creates a schedule that runs `ReportWorkflow` twice, one second
apart, waits for it to complete, lists and inspects its runs, then creates a
second schedule and triggers a run of it by hand. It deletes both schedules and
purges every run on the way out.

## Step

### Prepare

- Dapr 1.20 or later (dapr master until 1.20.0-rc.1 is released). Workflow
  schedules are not in any earlier release.
- A state store with actor support that can list keys, for `ListInstanceIDs`.
  The Redis store in `./config` does.

### Run Workflow Schedule

<!-- STEP
name: Run Workflow Schedule
output_match_mode: substring
expected_stdout_lines:
  - 'ReportWorkflow registered'
  - 'workflow worker started'
  - 'schedule created: report-every-1s'
  - 'schedule completed: runs=2'
  - 'last run status: COMPLETED'
  - 'listed 2 runs'
  - 'run output: "report for eu"'
  - 'schedule listed: true'
  - 'triggered run completed: "report for us"'
  - 'manual schedule run count: 1'
  - 'schedules deleted'
  - 'runs purged'
  - 'workflow schedule example completed'

background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id workflow-schedule \
         --dapr-grpc-port 50001 \
         --log-level debug \
         --resources-path ./config \
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

```
ReportWorkflow registered
workflow worker started
schedule created: report-every-1s
schedule completed: runs=2
last run status: COMPLETED
listed 2 runs
run output: "report for eu"
schedule listed: true
triggered run completed: "report for us"
manual schedule run count: 1
schedules deleted
runs purged
workflow schedule example completed
```

## Notes

- The schedule string is what the Dapr Jobs API accepts: a 6-field cron
  expression with seconds, `@every <duration>`, or `@daily` and the like.
- **Every run is a new workflow instance** with the ID
  `<name>_<owner>_<fire time>`, or `<name>_<owner>_<ms>_manual` for a
  triggered run, where `<owner>` tags the app that owns the schedule. List
  them with `ListInstanceIDs` and the schedule name as prefix.
- `WithScheduleRepeats` counts scheduled and triggered runs alike. The
  schedule completes when its last run starts, so poll the last run's status
  to know when that run finished.
- **Creating a schedule that already exists fails with `AlreadyExists`** unless
  `WithScheduleOverwrite` is set. `GetWorkflowSchedule`,
  `DeleteWorkflowSchedule` and `TriggerWorkflowSchedule` of a missing schedule
  fail with `NotFound`, and triggering a schedule with no runs left fails with
  `FailedPrecondition`.
- By default a schedule starts one run for each time it missed while the app
  was down. `WithScheduleSkipMissed` collapses them into one run.
- Deleting a schedule leaves the runs it already started untouched.
