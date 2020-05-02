# dapr SDK for Go

This is the dapr SDK (client) for Go.

## Installation

```
go get github.com/dapr/go-sdk
```

## Usage

The `example` folder contains a dapr enabled app that receives events (serving), and a client app that uses this SDK to invoke dapr API (client).

1. Run the serving app

```
cd example/serving
dapr run --app-id serving --protocol grpc --app-port 4000 go run main.go
```

2. Run the caller

```
cd example/client
dapr run --app-id caller go run main.go
```
