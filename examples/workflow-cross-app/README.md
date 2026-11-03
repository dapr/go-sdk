# Dapr Cross-App Workflow Operations Example with go-sdk

This example runs two Dapr apps:

- `host`: registers `CrossAppWorkflow` and its activities and runs the workflow
  worker. It owns every workflow instance created in this example.
- `client`: registers nothing and runs no worker. It drives the full workflow
  lifecycle on the host app by naming the host's app ID on each call.

Every client call uses one of the cross-app options from the
`github.com/dapr/durabletask-go/workflow` package:

| Operation                | Option                     |
| ------------------------ | -------------------------- |
| `ScheduleWorkflow`       | `WithAppID`                |
| `FetchWorkflowMetadata`  | `WithFetchAppID`           |
| `WaitForWorkflowStart`   | `WithFetchAppID`           |
| `WaitForWorkflowCompletion` | `WithFetchAppID`        |
| `RaiseEvent`             | `WithRaiseEventAppID`      |
| `SuspendWorkflow`        | `WithSuspendAppID`         |
| `ResumeWorkflow`         | `WithResumeAppID`          |
| `TerminateWorkflow`      | `WithTerminateAppID`       |
| `PurgeWorkflowState`     | `WithPurgeAppID`           |
| `RerunWorkflowFromEvent` | `WithRerunAppID`           |

These options select the target app only. There is deliberately no namespace
option at the client level: cross-app operations are always resolved within the
caller's own namespace.

The workflow parks on an external event, so suspending, resuming, raising an
event and terminating it are all meaningful while it is in flight.

## Step

### Prepare

- Dapr installed (`dapr init`, which provides the default placement and
  scheduler services, so no placement flags are needed below)

Both apps share the same `--resources-path ./config`, which declares a single
Redis state store with `actorStateStore` set to `"true"`. Neither app pins a
Dapr port: `dapr run` injects `DAPR_GRPC_PORT` per child process, so each app
automatically talks to its own sidecar.

### Run the workflow host app

<!-- STEP
name: Run workflow host app
background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id cross-app-host \
         --log-level debug \
         --resources-path ./config \
         -- go run ./host
```

<!-- END_STEP -->

### Run the workflow client app

<!-- STEP
name: Run workflow client app
output_match_mode: substring
expected_stdout_lines:
  - 'client: scheduled workflow cross-app-workflow-1 on app cross-app-host'
  - 'client: workflow cross-app-workflow-1 running on app cross-app-host'
  - 'client: workflow cross-app-workflow-1 suspended on app cross-app-host'
  - 'client: workflow cross-app-workflow-1 resumed on app cross-app-host'
  - 'client: raised event approval on workflow cross-app-workflow-1 on app cross-app-host'
  - 'client: workflow cross-app-workflow-1 completed on app cross-app-host with output: confirmed: hello dapr, approved'
  - 'client: workflow cross-app-workflow-1 purged on app cross-app-host'
  - 'client: scheduled workflow cross-app-workflow-2 on app cross-app-host'
  - 'client: workflow cross-app-workflow-2 running on app cross-app-host'
  - 'client: workflow cross-app-workflow-2 terminated on app cross-app-host'
  - 'client: workflow cross-app-workflow-2 purged on app cross-app-host'
  - 'client: all cross-app operations completed'
timeout_seconds: 120
-->

```bash
dapr run --app-id cross-app-client \
         --log-level debug \
         --resources-path ./config \
         -- go run ./client
```

<!-- END_STEP -->

### Cleanup

<!-- STEP
name: Cleanup workflow host app
expected_return_code:
-->

```bash
dapr stop --app-id cross-app-host
```

<!-- END_STEP -->

## Result

The client app prints:

```
client: scheduled workflow cross-app-workflow-1 on app cross-app-host
client: workflow cross-app-workflow-1 running on app cross-app-host
client: workflow cross-app-workflow-1 suspended on app cross-app-host
client: workflow cross-app-workflow-1 resumed on app cross-app-host
client: raised event approval on workflow cross-app-workflow-1 on app cross-app-host
client: workflow cross-app-workflow-1 completed on app cross-app-host with output: confirmed: hello dapr, approved
client: workflow cross-app-workflow-1 purged on app cross-app-host
client: scheduled workflow cross-app-workflow-2 on app cross-app-host
client: workflow cross-app-workflow-2 running on app cross-app-host
client: workflow cross-app-workflow-2 terminated on app cross-app-host
client: workflow cross-app-workflow-2 purged on app cross-app-host
client: all cross-app operations completed
```

The host app prints its registration lines and then stays quiet, since it never
initiates anything itself:

```
host: CrossAppWorkflow registered
host: GreetActivity registered
host: ConfirmActivity registered
host: worker started, waiting for cross-app operations
```

Host output is not asserted on above because the two processes run
concurrently and their output interleaves unpredictably.
