# Dapr gRPC Actor Example with go-sdk

This example hosts actors over the app-initiated actor event stream
(`SubscribeActorEventsAlpha1`): the application dials the Dapr sidecar's gRPC
port and receives all actor callbacks (method invocations, reminders, timers,
deactivations) over a single bidirectional stream, instead of exposing HTTP
actor endpoints.

> This feature is in **Alpha** and requires the Dapr sidecar to run with
> `--app-protocol grpc`.

The actor implementation is identical to the [HTTP actor example](../actor):
only the hosting changes, from `service/http` to
`client.SubscribeActorEvents`.

## Step

### Prepare

- Dapr installed

### Run Actor Server

<!-- STEP
name: Run Actor server
output_match_mode: substring
expected_stdout_lines:
  - 'actor event subscription started'
  - 'call get user req =  &{abc 123}'
  - 'get req =  laurence'
  - 'get post request =  laurence'
  - 'get req =  hello'
  - 'get req =  hello'
  - 'receive reminder =  testReminderName  state =  "hello"'
  - 'receive reminder =  testReminderName  state =  "hello"'
background: true
timeout_seconds: 60
-->

```bash
dapr run --app-id actor-grpc-serving \
         --app-protocol grpc \
         --app-port 50051 \
         --log-level debug \
         --resources-path ./config \
         go run ./serving/main.go
```

<!-- END_STEP -->

### Run Actor Client

<!-- STEP
name: Run Actor Client
output_match_mode: substring
expected_stdout_lines:
  - 'get user result =  &{abc 123}'
  - 'get invoke result =  laurence'
  - 'get post result =  laurence'
  - 'get result =  get result'
  - 'start timer'
  - 'stop timer'
  - 'start reminder'
  - 'stop reminder'
  - 'get user = {Name: Age:1}'
  - 'get user = {Name: Age:2}'

timeout_seconds: 60
-->

```bash
dapr run --app-id actor-grpc-client \
         --log-level debug \
         --resources-path ./config \
         go run ./client/main.go
```

<!-- END_STEP -->

### Cleanup

<!-- STEP
name: Cleanup actor server
expected_return_code:
-->

```bash
dapr stop --app-id actor-grpc-serving
(lsof -i:50051 | grep main) | awk '{print $2}' | xargs kill
```

<!-- END_STEP -->

## Result
- client side
```
dapr client initializing for: 127.0.0.1:55776
get user result =  &{abc 123}
get invoke result =  laurence
get post result =  laurence
get result =  get result
start timer
stop timer
start reminder
stop reminder
get user = {Name: Age:1}
get user = {Name: Age:2}
✅  Exited App successfully
```

- server side

```
actor event subscription started
call get user req =  &{abc 123}
get req =  laurence
get post request =  laurence
get req =  hello
get req =  hello
receive reminder =  testReminderName  state =  "hello"
receive reminder =  testReminderName  state =  "hello"
```

## Notes

- The gRPC service on `:50051` only backs the sidecar's app channel (which
  must be gRPC for the stream to be accepted); actor callbacks never reach
  it — they are all delivered over the actor event stream.
- The stream automatically reconnects and re-registers the actor types if the
  sidecar restarts or the stream drops.
