module github.com/dapr/go-sdk/examples/configuration

go 1.19

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../

require (
	github.com/dapr/go-sdk v1.8.0
	github.com/go-redis/redis/v8 v8.11.5
	google.golang.org/grpc v1.55.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
