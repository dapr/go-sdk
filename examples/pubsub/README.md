# Dapr PubSub Example with go-sdk

This folder contains two Go files that use the Go SDK to invoke the Dapr Pub/Sub API.

## Diagram

![](https://i.loli.net/2020/08/23/5MBYgwqCZcXNUf2.jpg)

## Step

### Prepare

- Dapr installed

### Run Subscriber Server

<!-- STEP
name: Run Subscriber Server
output_match_mode: substring
expected_stdout_lines:
  - 'event - PubsubName: messages, Topic: neworder'
background: true
sleep: 15
-->

```bash
dapr run --app-id sub \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --components-path ./config \
         go run sub/sub.go
```

<!-- END_STEP -->

### Run Publisher

<!-- STEP
name: Run publisher
expected_stdout_lines:
  - '== APP == data published'
background: true
sleep: 15
-->

```bash
export DAPR_PUBSUB_NAME=messages

dapr run --app-id pub \
         --log-level debug \
         --components-path ./config \
         go run pub/pub.go
```

<!-- END_STEP -->

### Cleanup

<!-- STEP
expected_stdout_lines: 
  - '✅  app stopped successfully: sub'
expected_stderr_lines:
name: Shutdown dapr
-->

```bash
dapr stop --app-id sub
(lsof -i:8080 | grep sub) | awk '{print $2}' | xargs  kill
```

<!-- END_STEP -->

## Result

```shell
== APP == 2020/08/23 13:21:58 event - PubsubName: messages, Topic: neworder, ID: 11acaa82-23c4-4244-8969-7360dae52e5d, Data: ping
```
