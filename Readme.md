# Dapr SDK for Go

This is the dapr SDK (client) for Go.

## Installation

```
go get github.com/dapr/go-sdk
```

## Usage

The `example` folder contains a Dapr enabled `serving` app a `client` app that uses this SDK to invoke Dapr API for state and events, `serving` app for service to service invocation, and a simple HTTP binding to illustrate output binding. To run the example:

1. Start the `serving` app in the `example/serving` directory 

```
cd example/serving
dapr run --app-id serving --protocol grpc --app-port 50001 go run main.go
```

2. Start the `client` app in the `example/client` directory

```
cd example/client
dapr run --app-id caller go run main.go --components-path ./components
```
