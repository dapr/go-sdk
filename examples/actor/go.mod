module github.com/dapr/go-sdk/examples/actor

go 1.21
toolchain go1.22.2

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../

require (
	github.com/dapr/go-sdk v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require (
	github.com/dapr/dapr v1.13.0 // indirect
	github.com/go-chi/chi/v5 v5.0.12 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	go.opentelemetry.io/otel v1.23.1 // indirect
	go.opentelemetry.io/otel/trace v1.23.1 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240205150955-31a09d347014 // indirect
	google.golang.org/grpc v1.62.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
