module github.com/dapr/go-sdk

go 1.26.6

require (
	github.com/dapr/dapr v1.18.0
	github.com/dapr/durabletask-go v0.14.2-0.20260923181807-3ab246ba418d
	github.com/dapr/kit v0.18.3-0.20260727141402-dd127582d044
	github.com/go-chi/chi/v5 v5.2.4
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.6.0
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spiffe/go-spiffe/v2 v2.7.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260727163830-6c54dddc4772 // indirect
)

replace github.com/dapr/dapr => github.com/mikeee/dapr v1.12.1-0.20260924095259-134057580d3e
