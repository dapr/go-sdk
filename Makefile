RELEASE_VERSION  =v0.10.0
GDOC_PORT        =8888
PROTO_ROOT       =https://raw.githubusercontent.com/dapr/dapr/master/dapr/proto/

.PHONY: mod test cover service client lint protps tag docs clean help
all: test

tidy: ## Updates the go modules
	go mod tidy

test: mod ## Tests the entire project 
	go test -count=1 -race ./...

cover: mod ## Displays test coverage in the client and service packages
	go test -coverprofile=cover-client.out ./client && go tool cover -html=cover-client.out
	go test -coverprofile=cover-grpc.out ./service/grpc && go tool cover -html=cover-grpc.out
	go test -coverprofile=cover-http.out ./service/http && go tool cover -html=cover-http.out

service: mod ## Runs the uncompiled example service code 
	dapr run --app-id serving \
			 --app-protocol grpc \
			 --app-port 50001 \
			 --port 3500 \
			 --log-level debug \
			 --components-path example/serving/grpc/config \
			 go run example/serving/grpc/main.go

service-v09http: mod ## Runs the uncompiled HTTP example service code using the Dapr v0.9 flags
	dapr run --app-id serving \
			 --protocol http \
			 --app-port 8080 \
			 --port 3500 \
			 --log-level debug \
			 --components-path example/serving/http/config \
			 go run example/serving/http/main.go

	: mod ## Runs the uncompiled gRPC example service code using the Dapr v0.9 flags
	dapr run --app-id serving \
			 --protocol grpc \
			 --app-port 50001 \
			 --port 3500 \
			 --log-level debug \
			 --components-path example/serving/grpc/config \
			 go run example/serving/grpc/main.go

client: mod ## Runs the uncompiled example client code 
	dapr run --app-id caller \
             --components-path example/client/config \
             --log-level debug \
             go run example/client/main.go 

pubsub: ## Submits pub/sub events in different cotnent types 
	curl -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
		-H "Content-type: application/json" \
		"http://localhost:3500/v1.0/publish/messages"
	curl -d '<message><from>John</from><to>Lary</to></message>' \
		-H "Content-type: application/xml" \
		"http://localhost:3500/v1.0/publish/messages"
	curl -d '0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40' \
		-H "Content-type: application/octet-stream" \
		"http://localhost:3500/v1.0/publish/messages"

invoke: ## Invokes service method with different operations
	curl -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
		-H "Content-type: application/json" \
		"http://localhost:3500/v1.0/invoke/serving/method/echo"
	curl -d "ping" \
		-H "Content-type: text/plain;charset=UTF-8" \
		"http://localhost:3500/v1.0/invoke/serving/method/echo"
	curl -X DELETE \
		"http://localhost:3500/v1.0/invoke/serving/method/echo?k1=v1&k2=v2"

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

protos: ## Downloads proto files from dapr/dapr master and generats gRPC proto clients
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
