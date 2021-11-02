# Dapr Go client example

The `examples/service` folder contains a Dapr enabled `serving` app and a `client` app that uses this SDK to invoke Dapr API for state and events, The `serving` app is available as HTTP or gRPC. The `client` app can target either one of these for service to service and binding invocations.

To run this example, start by first launching the service in either HTTP or gRPC:

### Prepare

- Dapr installed

### HTTP

<!-- STEP
name: Run Subscriber Server
output_match_mode: substring
expected_stdout_lines:
  - "ContentType:text/plain, Verb:POST, QueryString:, hellow"
background: true
sleep: 15
-->

```bash
dapr run --app-id serving \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --components-path ./config \
         go run ./serving/http/main.go
```

<!-- END_STEP -->

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

<!-- STEP
name: Run publisher
expected_stdout_lines:
  - '== APP == data published'
  - '== APP == saving data: { "message": "hello" }'
  - '== APP == data saved'
  - '== APP == data retrieved [key:key1 etag:1]: { "message": "hello" }'
  - '== APP == data item saved'
  - '== APP == data deleted'
  - '== APP == service method invoked, response: hellow'
  - '== APP == output binding invoked'
background: true
sleep: 15
-->

```bash
dapr run --app-id caller \
         --components-path ./config \
         --log-level debug \
         go run ./client/main.go
```

<!-- END_STEP -->

## Custom gRPC client

Launch the DAPR client with custom gRPC client to accept and receive payload size > 4 MB:

<!-- STEP
output_match_mode: substring
expected_stdout_lines:
  - '== APP == Writing large data blob'
  - '== APP == Saved the large data blob'
  - '== APP == Writing to statestore took'
  - '== APP == Reading from statestore took'
  - '== APP == Deleting key from statestore took'
  - '== APP == DONE (CTRL+C to Exit)'
-->

```bash
dapr run --app-id custom-grpc-client \
		 -d ./config \
		 --dapr-http-max-request-size 41 \
		 --log-level debug \
		 go run ./custom-grpc-client/main.go
```

<!-- END_STEP -->

## API

### PubSub

Publish JSON content

```shell
curl -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
     -H "Content-type: application/json" \
     "http://localhost:3500/v1.0/publish/messages/topic1"
```

Publish XML content (read as text)

```shell
curl -d '<message><from>John</from><to>Lary</to></message>' \
     -H "Content-type: application/xml" \
     "http://localhost:3500/v1.0/publish/messages/topic1"
```

Publish BIN content

```shell
curl -d '0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40' \
     -H "Content-type: application/octet-stream" \
     "http://localhost:3500/v1.0/publish/messages/topic1"
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

### Cleanup

<!-- STEP
expected_stdout_lines: 
  - '✅  app stopped successfully: serving'
expected_stderr_lines:
name: Shutdown dapr
-->

```bash
dapr stop --app-id serving
(lsof -i:8080 | grep main) | awk '{print $2}' | xargs  kill
```

<!-- END_STEP -->

