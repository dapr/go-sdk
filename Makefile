RELEASE_VERSION  =v1.0.0-rc-3
GDOC_PORT        =8888
PROTO_ROOT       =https://raw.githubusercontent.com/dapr/dapr/master/dapr/proto/
GO_COMPAT_VERSION=1.18

.PHONY: all
all: help

.PHONY: tidy
tidy: ## Updates the go modules
	go mod tidy -compat=$(GO_COMPAT_VERSION)

.PHONY: test
test:
	go test -count=1 \
			-race \
			-coverprofile=coverage.txt \
			-covermode=atomic \
			./...

.PHONY: spell
spell: ## Checks spelling across the entire project 
	@command -v misspell > /dev/null 2>&1 || (cd tools && go get github.com/client9/misspell/cmd/misspell)
	@misspell -locale US -error go=golang client/**/* examples/**/* service/**/* actor/**/* .

.PHONY: cover
cover: ## Displays test coverage in the client and service packages
	go test -coverprofile=cover-client.out ./client && go tool cover -html=cover-client.out
	go test -coverprofile=cover-grpc.out ./service/grpc && go tool cover -html=cover-grpc.out
	go test -coverprofile=cover-http.out ./service/http && go tool cover -html=cover-http.out

.PHONY: lint
lint: ## Lints the entire project
	golangci-lint run --timeout=3m

.PHONY: tag
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean
clean: ## Cleans go and generated files in ./dapr/proto/
	go clean
	rm -fr ./dapr/proto/common/v1/*
	rm -fr ./dapr/proto/runtime/v1/*

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
