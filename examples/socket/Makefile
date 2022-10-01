.PHONY: bin
bin: go-mod
	go build -o order order.go

.PHONY: go-mod
go-mod: go-check
	go mod tidy -compat=1.17
	go mod vendor

.PHONY: go-check
go-check:
	@which go > /dev/null

.PHONY: clean
clean:
	rm -f ./order
