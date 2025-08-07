module github.com/dapr/go-sdk/examples

go 1.24.4

replace github.com/dapr/go-sdk => ../

require (
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/dapr/go-sdk v0.0.0-00010101000000-000000000000
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.73.0
	google.golang.org/grpc/examples v0.0.0-20240516203910-e22436abb809
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dapr/dapr v1.16.0-rc.3 // indirect
	github.com/dapr/durabletask-go v0.7.3-0.20250711135247-7a35af6fe0e5 // indirect
	github.com/dapr/kit v0.15.4 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/utils v0.0.0-20250502105355-0f33e8f1c979 // indirect
)
