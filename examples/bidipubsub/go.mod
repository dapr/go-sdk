module github.com/dapr/go-sdk/examples/pubsub

go 1.22.3

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../

require github.com/dapr/go-sdk v0.0.0-00010101000000-000000000000

require (
	github.com/dapr/dapr v1.13.0-rc.1.0.20240614011318-a04348b0099d // indirect
	github.com/go-chi/chi/v5 v5.0.12 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.62.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
