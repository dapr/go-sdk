module github.com/dapr/go-sdk/examples/actor

go 1.19

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../

require (
	github.com/dapr/go-sdk v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.3.1
)

require (
	github.com/dapr/dapr v1.12.0-rc.4 // indirect
	github.com/go-chi/chi/v5 v5.0.10 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230807174057-1744710a1577 // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
