module github.com/dapr/go-sdk/examples/pubsub

go 1.17

require github.com/dapr/go-sdk v1.3.1-0.20211214200612-a38be4e38b7d

require (
	github.com/dapr/dapr v1.6.0-rc.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2 // indirect
	google.golang.org/grpc v1.40.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../
