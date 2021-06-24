# Dapr PubSub Example with go-sdk

This folder contains two Go files that use the Go SDK to invoke the Dapr Pub/Sub API.

## Diagram

![](https://i.loli.net/2020/08/23/5MBYgwqCZcXNUf2.jpg)

## Step

### Prepare

- Dapr installed

### Run Subscriber Server

```shell
dapr run --app-id sub \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --components-path ./config \
         go run sub/sub.go
```

### Run Publisher

```shell
export DAPR_PUBSUB_NAME=messages

dapr run --app-id pub \
         --log-level debug \
         --components-path ./config \
         go run pub/pub.go
```

## Result

```shell
== APP == 2020/08/23 13:21:58 event - PubsubName: messages, Topic: neworder, ID: 11acaa82-23c4-4244-8969-7360dae52e5d, Data: ping
```
