GO_COMPAT_VERSION=1.22

.PHONY: cover
cover:
	go test -coverprofile=cover.out ./ && go tool cover -html=cover.out

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
