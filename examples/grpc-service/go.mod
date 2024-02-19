module github.com/dapr/go-sdk/examples/grpc-service

go 1.21

toolchain go1.21.6

replace github.com/dapr/go-sdk => ../../

require (
	github.com/dapr/go-sdk v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.61.0
	google.golang.org/grpc/examples v0.0.0-20240205234101-d41b01db97ca
)

require (
	github.com/dapr/dapr v1.13.0-rc.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	go.opentelemetry.io/otel v1.23.1 // indirect
	go.opentelemetry.io/otel/trace v1.23.1 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240205150955-31a09d347014 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)
