# Dapr SDK for Go

This is the Dapr SDK for Go, based on the auto-generated proto client.

## Installation

```
go get github.com/dapr/go-sdk
```

## Usage

The `example` folder contains a Dapr enabled app that receives events (client), and a caller that invokes the Dapr API (caller).

1. Run the client

```
cd example/client
dapr run --app-id client --protocol grpc --app-port 4000 go run main.go
```

2. Run the caller

```
cd example/caller
dapr run --app-id caller go run main.go
```

*Note: If you don't setup a Dapr binding, expect the error message `rpc error: code = Unknown desc = ERR_INVOKE_OUTPUT_BINDING: couldn't find output binding storage`*