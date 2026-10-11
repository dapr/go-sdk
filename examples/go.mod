module github.com/dapr/go-sdk/examples

go 1.26.9

replace github.com/dapr/go-sdk => ../

require (
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/dapr/durabletask-go v0.15.0
	github.com/dapr/go-sdk v0.0.0-00010101000000-000000000000
	github.com/dapr/kit v0.18.3-0.20261002162430-f2a7561ba0de
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.83.2
	google.golang.org/grpc/examples v0.0.0-20250407062114-b368379ef8f6
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dapr/dapr v1.19.0-rc.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-chi/chi/v5 v5.3.2 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/spiffe/go-spiffe/v2 v2.7.0 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260803160001-6ac0973c030d // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
