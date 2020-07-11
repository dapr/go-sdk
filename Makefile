RELEASE_VERSION  =v0.8.0-r1
GDOC_PORT        =8888
PROTO_ROOT       =https://raw.githubusercontent.com/dapr/dapr/master/dapr/proto/

.PHONY: mod test cover service client lint protps tag docs clean help
all: test

mod: ## Updates the go modules
	go mod tidy

test: mod ## Tests the entire project 
	go test -v -count=1 -race ./...

cover: mod ## Displays test coverage in the Client package
	go test -coverprofile=cover.out ./client && go tool cover -html=cover.out

service: mod ## Runs the uncompiled example service code 
	dapr run --app-id serving \
	         --protocol grpc \
			 --app-port 50001 \
			 go run example/serving/main.go

client: mod ## Runs the uncompiled example client code 
	dapr run --app-id caller \
             --components-path example/client/comp \
             go run example/client/main.go 

lint: ## Lints the entire project
	golangci-lint run --timeout=3m

docs: ## Runs godoc (in container due to mod support)
	docker run \
			--rm \
			-e "GOPATH=/tmp/go" \
			-p 127.0.0.1:$(GDOC_PORT):$(GDOC_PORT) \
			-v $(PWD):/tmp/go/src/ \
			--name godoc golang \
			bash -c "go get golang.org/x/tools/cmd/godoc && echo http://localhost:$(GDOC_PORT)/pkg/ && /tmp/go/bin/godoc -http=:$(GDOC_PORT)"
	open http://localhost:8888/pkg/client/

tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

clean: ## Cleans go and generated files in ./dapr/proto/
	go clean
	rm -fr ./dapr/proto/common/v1/*
	rm -fr ./dapr/proto/runtime/v1/*

protos: ## Downloads proto files from dapr/dapr and generats gRPC proto clients
	go install github.com/gogo/protobuf/gogoreplace

	wget -q $(PROTO_ROOT)/common/v1/common.proto -O ./dapr/proto/common/v1/common.proto
	gogoreplace 'option go_package = "github.com/dapr/dapr/pkg/proto/common/v1;common";' \
		'option go_package = "github.com/dapr/go-sdk/dapr/proto/common/v1;common";' \
		./dapr/proto/common/v1/common.proto

	wget -q $(PROTO_ROOT)/runtime/v1/appcallback.proto -O ./dapr/proto/runtime/v1/appcallback.proto
	gogoreplace 'option go_package = "github.com/dapr/dapr/pkg/proto/runtime/v1;runtime";' \
		'option go_package = "github.com/dapr/go-sdk/dapr/proto/runtime/v1;runtime";' \
		./dapr/proto/runtime/v1/appcallback.proto

	wget -q $(PROTO_ROOT)/runtime/v1/dapr.proto -O ./dapr/proto/runtime/v1/dapr.proto
	gogoreplace 'option go_package = "github.com/dapr/dapr/pkg/proto/runtime/v1;runtime";' \
		'option go_package = "github.com/dapr/go-sdk/dapr/proto/runtime/v1;runtime";' \
		./dapr/proto/runtime/v1/dapr.proto

	protoc -I . --go_out=plugins=grpc:. --go_opt=paths=source_relative  ./dapr/proto/common/v1/*.proto
	protoc -I . --go_out=plugins=grpc:. --go_opt=paths=source_relative ./dapr/proto/runtime/v1/*.proto

	rm -f ./dapr/proto/common/v1/*.proto
	rm -f ./dapr/proto/runtime/v1/*.proto

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
