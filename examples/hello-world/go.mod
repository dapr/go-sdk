module github.com/dapr/go-sdk/examples/hello-world

go 1.17

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/dapr/go-sdk v1.3.1-0.20211214200612-a38be4e38b7d
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

require (
	github.com/dapr/dapr v1.7.4-0.20220620022343-b22c67f67b3c // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220405205423-9d709892a2bf // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// Needed to validate SDK changes in CI/CD
replace github.com/dapr/go-sdk => ../../
