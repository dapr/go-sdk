# Grpc Service Example with proxy mode

The `examples/grpc-service` folder contains a Dapr enabled `server` app and a `client` app that uses this SDK to invoke grpc methos via grpc stub, The `server` app is available as gRPC. The `client` app can target either one of these for service to service and binding invocations.


## Step

### Prepare

- Dapr installed

### Run server as a dapr app

<!-- STEP
name: Run grpc server with dapr proxy mode
output_match_mode: substring
expected_stdout_lines:
  - 'Received: Dapr'
background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id grpc-server \
         --app-port 50051 \
         --app-protocol grpc \
         --dapr-grpc-port 50007 \
         go run ./server/main.go
```

<!-- END_STEP -->

### Run grpc client

<!-- STEP
name: Run grpc client
expected_stdout_lines:
  - 'Greeting: Hello Dapr'
output_match_mode: substring
background: true
sleep: 15
timeout_seconds: 60
-->

```bash
dapr run --app-id grpc-client \
         go run ./client/main.go
```

<!-- END_STEP -->

### Cleanup

```bash
dapr stop --app-id grpc-server
```
