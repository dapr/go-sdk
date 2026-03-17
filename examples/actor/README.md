# Dapr Actor Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Actor Server

<!-- STEP
name: Run Actor server
output_match_mode: substring
expected_stdout_lines:
  - 'call get user req =  &{abc 123}'
  - 'get req =  laurence'
  - 'get post request =  laurence'
  - 'get req =  hello'
  - 'get req =  hello'
  - 'receive reminder =  testReminderName  state =  "hello"'
  - 'receive reminder =  testReminderName  state =  "hello"'
background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id actor-serving \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
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

background: true
sleep: 40
timeout_seconds: 60
-->

```bash
dapr run --app-id actor-client \
         --log-level debug \
         --resources-path ./config \
         go run ./client/main.go
```

<!-- END_STEP -->

### Cleanup

```bash
dapr stop --app-id  actor-serving
(lsof -i:8080 | grep main) | awk '{print $2}' | xargs  kill
```

## Result
- client side
```
== APP == dapr client initializing for: 127.0.0.1:55776
== APP == get user result =  &{abc 123}
== APP == get invoke result =  laurence
== APP == get post result =  laurence
== APP == get result =  get result
== APP == start timer
== APP == stop timer
== APP == start reminder
== APP == stop reminder
== APP == get user = {Name: Age:1}
== APP == get user = {Name: Age:2}
✅  Exited App successfully
```

- server side

```
== APP == call get user req =  &{abc 123}
== APP == get req =  laurence
== APP == get post request =  laurence
== APP == get req =  hello
== APP == get req =  hello
== APP == receive reminder =  testReminderName  state =  "hello"
== APP == receive reminder =  testReminderName  state =  "hello"
```
