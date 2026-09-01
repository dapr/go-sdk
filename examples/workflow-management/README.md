# Dapr Workflow Management Example with go-sdk

Demonstrates the advanced workflow management APIs exposed by the Dapr Go SDK:

| API | What it does |
| --- | --- |
| `ListInstanceIDs` | Enumerate the app's workflow instance IDs, with pagination |
| `GetInstanceHistory` | Read an instance's full execution history |
| `RerunWorkflowFromEvent` | Rerun ("rewind") an instance from one of its history events |

The three compose naturally, and the example uses them that way: list the
instances, read one instance's history to find an event worth rerunning from,
then rerun from that event with a fresh input.

## Step

### Prepare

- Dapr 1.18 or later. All three APIs were added in 1.18, and `ListInstanceIDs`
  additionally needs a state store that supports key listing.
- A state store with actor support that can list keys. The Redis store in
  `./config` does.

### Run Workflow Management

<!-- STEP
name: Run Workflow Management
output_match_mode: substring
expected_stdout_lines:
  - 'OrderWorkflow registered'
  - 'workflow worker started'
  - 'scheduled and completed 3 workflows'
  - '== ListInstanceIDs =='
  - 'listed 3 instances with prefix order- over 2 page(s)'
  - '  instance: order-01'
  - '  instance: order-02'
  - '  instance: order-03'
  - '== GetInstanceHistory =='
  - 'rerunnable event: type=TaskScheduled name=ReserveStock'
  - 'rerunnable event: type=TaskScheduled name=ChargeCard'
  - '== RerunWorkflowFromEvent =='
  - 'rerunning order-01 from its ChargeCard activity as order-01-rerun'
  - 'rerun instance order-01-rerun finished with status: COMPLETED, output: "charged 25"'
  - 'purged all instances'
  - 'history for purged instance is gone'
  - 'workflow management example completed'

background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id workflow-management \
         --dapr-grpc-port 50001 \
         --log-level debug \
         --resources-path ./config \
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

```
OrderWorkflow registered
workflow worker started
scheduled and completed 3 workflows
== ListInstanceIDs ==
listed 3 instances with prefix order- over 2 page(s)
  instance: order-01
  instance: order-02
  instance: order-03
== GetInstanceHistory ==
instance order-01 has 9 history events
  rerunnable event: type=TaskScheduled name=ReserveStock
  rerunnable event: type=TaskScheduled name=ChargeCard
== RerunWorkflowFromEvent ==
rerunning order-01 from its ChargeCard activity as order-01-rerun
rerun instance order-01-rerun finished with status: COMPLETED, output: "charged 25"
purged all instances
history for purged instance is gone
workflow management example completed
```

The history event count depends on turn boundaries, so it is not asserted above.
The page count is asserted: with three instances and a page size of two, Redis
returns exactly two pages.

## Notes

- **Rerun requires a terminal source instance.** Rerunning an instance that is
  still running fails with `InvalidArgument: '<id>' is not in a terminal state`.
- **The source cannot be a child workflow**, which also fails with
  `InvalidArgument`.
- **Only three event types can be rerun**: `TaskScheduled` (a scheduled
  activity), `TimerCreated` and `ChildWorkflowInstanceCreated`. Any other event
  is rejected with `NotFound: target event '<type>' with ID '<n>' is not an
  event that can be rerun`.
- **Target an event's own `EventId`**, not its position in the history slice.
- `WithRerunInput` replaces the input of the target event only. Work that
  completed and was recorded before the target event is replayed from history,
  but anything from before it that had not completed (in-flight activities,
  unfired timers, unfinished child workflows) is re-executed. In a parallel
  workflow this includes an activity whose completion was recorded after the
  target event.
- `WithRerunInput` is invalid on a timer event, and `WithRerunNewChildInstanceID`
  is only valid on a child-workflow event.
