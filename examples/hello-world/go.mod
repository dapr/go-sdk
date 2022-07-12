module github.com/dapr/go-sdk/examples/hello-world

go 1.18

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/dapr/go-sdk v1.3.1-0.20211214200612-a38be4e38b7d
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

require (
	github.com/dapr/dapr v1.8.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20220621193019-9d032be2e588 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220622171453-ea41d75dfa0f // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../
