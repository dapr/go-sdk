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
match_order: none
expected_stdout_lines:
  - 'event - PubsubName: messages, Topic: neworder'
  - 'event - PubsubName: messages, Topic: neworder'
  - 'event - PubsubName: messages, Topic: neworder'
  - 'event - PubsubName: messages, Topic: sendorder'
  - 'event - PubsubName: messages, Topic: sendorder'
  - 'event - PubsubName: messages, Topic: sendorder'
expected_stderr_lines:
background: true
sleep: 15
-->

```bash
dapr run --app-id sub \
         --dapr-http-port 3500 \
         --log-level debug \
         --resources-path ./config \
         go run sub/sub.go
```

<!-- END_STEP -->

### Run Publisher

<!-- STEP
name: Run publisher
output_match_mode: substring
expected_stdout_lines:
  - 'sending message'
  - 'message published'
  - 'sending multiple messages'
  - 'multiple messages published'
expected_stderr_lines:
background: true
sleep: 15
-->

```bash
dapr run --app-id pub \
         --log-level debug \
         --resources-path ./config \
         go run pub/pub.go
```

<!-- END_STEP -->

## Result

```shell
== APP == 2023/03/29 21:36:07 event - PubsubName: messages, Topic: neworder, ID: 82427280-1c18-4fab-b901-c7e68d295d31, Data: ping123
```
