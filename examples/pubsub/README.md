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
timeout_seconds: 60
-->

```bash
dapr run --app-id sub \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --resources-path ./config \
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
timeout_seconds: 60
-->

```bash
export DAPR_PUBSUB_NAME=messages

dapr run --app-id pub \
         --log-level debug \
         --resources-path ./config \
         go run pub/pub.go
```

<!-- END_STEP -->

### Cleanup

```bash
dapr stop --app-id sub
(lsof -i:8080 | grep sub) | awk '{print $2}' | xargs  kill
```

## Result

```shell
== APP == 2023/03/29 21:36:07 event - PubsubName: messages, Topic: neworder, ID: 82427280-1c18-4fab-b901-c7e68d295d31, Data: ping
== APP == 2023/03/29 21:36:07 event - PubsubName: messages, Topic: neworder, ID: cc13829c-af77-4303-a4d7-55cdc0b0fa7d, Data: multi-pong
== APP == 2023/03/29 21:36:07 event - PubsubName: messages, Topic: neworder, ID: 0147f10a-d6c3-4b16-ad5a-6776956757dd, Data: multi-ping
```
