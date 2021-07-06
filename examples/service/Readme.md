# Dapr Go client example

The `examples/service` folder contains a Dapr enabled `serving` app and a `client` app that uses this SDK to invoke Dapr API for state and events, The `serving` app is available as HTTP or gRPC. The `client` app can target either one of these for service to service and binding invocations.

To run this example, start by first launching the service in either HTTP or gRPC:

### HTTP

```
dapr run --app-id serving \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --components-path ./config \
         go run ./serving/http/main.go
```

### gRPC

```
dapr run --app-id serving \
         --app-protocol grpc \
         --app-port 50001 \
         --dapr-grpc-port 3500 \
         --log-level debug \
         --components-path ./config \
         go run ./serving/grpc/main.go
```

## Client 

Once one of the above services is running, launch the client:

```
dapr run --app-id caller \
         --components-path ./config \
         --log-level debug \
         go run ./client/main.go
```

## API

### PubSub

Publish JSON content

```shell
curl -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
     -H "Content-type: application/json" \
     "http://localhost:3500/v1.0/publish/messages"
```

Publish XML content (read as text)

```shell
curl -d '<message><from>John</from><to>Lary</to></message>' \
     -H "Content-type: application/xml" \
     "http://localhost:3500/v1.0/publish/messages"
```

Publish BIN content 

```shell
curl -d '0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40' \
     -H "Content-type: application/octet-stream" \
     "http://localhost:3500/v1.0/publish/messages"
```

### Service Invocation 

Invoke service with JSON payload

```shell
curl -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
     -H "Content-type: application/json" \
     "http://localhost:3500/v1.0/invoke/serving/method/echo"
```

Invoke service with plain text message

```shell
curl -d "ping" \
     -H "Content-type: text/plain;charset=UTF-8" \
     "http://localhost:3500/v1.0/invoke/serving/method/echo"
```

Invoke service with no content

```shell
curl -X DELETE \
    "http://localhost:3500/v1.0/invoke/serving/method/echo?k1=v1&k2=v2"
```

### Input Binding  

Uses the [config/cron.yaml](config/cron.yaml) component