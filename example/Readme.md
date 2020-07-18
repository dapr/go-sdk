# Dapr go client example 


The `example` folder contains a Dapr enabled `serving` app and a `client` app that uses this SDK to invoke Dapr API for state and events, The `serving` app is available as HTTP or gRPC. The `client` app can target either one of these for service to service and binding invocations.

To run this example, start by first launching either `gRPC` or `HTTP` service:

## gRPC Service 

```
cd example/serving/grpc
dapr run --app-id serving \
         --protocol grpc \
         --app-port 50001 \
         --log-level debug \
         --components-path ./config \
         go run main.go
```

## HTTP Service 

```
cd example/serving/http
dapr run --app-id serving \
         --protocol http \
         --port 3500 \
         --app-port 8080 \
         --log-level debug \
         --components-path ./config \
         go run main.go
```

## Client 

Once one of the above services is running is running, launch the client:

```
cd example/client
dapr run --app-id caller \
         --components-path ./config \
         --log-level debug \
         go run main.go 
```