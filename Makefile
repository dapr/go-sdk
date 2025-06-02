RELEASE_VERSION  =v1.0.0-rc-3
GDOC_PORT        =8888
GO_COMPAT_VERSION=1.22

.PHONY: all
all: help

.PHONY: tidy
tidy: ## Updates the go modules
	go mod tidy -compat=$(GO_COMPAT_VERSION)

.PHONY: test
test:
	CGO_ENABLED=1 go test -count=1 \
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
	go test -coverprofile=cover-workflow.out ./workflow && go tool cover -html=cover-workflow.out

.PHONY: lint
lint: check-lint ## Lints the entire project
	golangci-lint run --timeout=3m

.PHONY: lint-fix
lint-fix: check-lint ## Lints the entire project
	golangci-lint run --timeout=3m --fix

.PHONY: check-lint
check-lint: ## Compares the locally installed linter with the workflow version
	cd ./tools/check-lint-version && \
	go mod tidy && \
	go run main.go

.PHONY: tag
tag: ## Creates release tag
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: check-diff
check-diff:
	git diff --exit-code ./go.mod # check no changes
	git diff --exit-code ./go.sum # check no changes

.PHONY: modtidy
modtidy:
	go mod tidy
