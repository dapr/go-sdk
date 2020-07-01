# Dapr go client example 


The `example` folder contains a Dapr enabled `serving` app a `client` app that uses this SDK to invoke Dapr API for state and events, `serving` app for service to service invocation, and a simple HTTP binding to illustrate output binding. 

To run the example first start the `gRPC` or `HTTP` server

## gRPC Server 

```
cd example/serving/grpc
dapr run --app-id serving \
         --protocol grpc \
         --app-port 50001 \
         go run main.go
```

## HTTP Server 

```
cd example/serving/http
dapr run --app-id serving \
         --protocol http \
         --app-port 8080 \
         go run main.go
```

## Client 


```
cd example/client
dapr run --app-id caller \
         --components-path ./config \
         go run main.go 
```
