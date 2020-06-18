RELEASE_VERSION  =v0.8.0
GDOC_PORT        =8888

.PHONY: mod test service client lint protps tag lint docs clean help
all: test

protos: ## Downloads proto files from dapr/dapr, generates gRPC clients
	bin/protogen

mod: ## Updates the go modules
	go mod tidy

test: mod ## Tests the entire project 
	go test -v -count=1 -race ./...
	# go test -v -count=1 -run TestInvokeServiceWithContent ./...

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

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
