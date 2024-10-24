all: help

.PHONY: http-service
http-service: ## Runs the uncompiled HTTP example service code
	dapr run --app-id serving \
			 -P http \
			 -p 8080 \
			 -H 3500 \
			 -d ./config \
			 --log-level debug \
			 go run ./serving/http/main.go

.PHONY: grpc-service
grpc-service: ## Runs the uncompiled gRPC example service code
	dapr run --app-id serving \
			 -P grpc \
			 -p 50001 \
			 -H 3500 \
			 -d ./config \
			 --log-level debug \
			 go run ./serving/grpc/main.go

.PHONY: client
client: ## Runs the uncompiled example client code 
	dapr run --app-id caller \
             -d ./config \
             --log-level debug \
             go run ./client/main.go 

.PHONY: custom-grpc-client
custom-grpc-client: ## Runs the uncompiled example custom grpc client code
	dapr run --app-id custom-grpc-client \
		 -d ./config \
		 --dapr-http-max-request-size 41 \
		 --log-level debug \
		 go run ./custom-grpc-client/main.go

.PHONY: publish
publish: ## Submits pub/sub events in different content types 
	curl -i -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
		-H "Content-type: application/json" \
		"http://localhost:3500/v1.0/publish/messages/topic1"
	curl -i -d '<message><from>John</from><to>Lary</to></message>' \
		-H "Content-type: application/xml" \
		"http://localhost:3500/v1.0/publish/messages/topic1"
	curl -i -d '0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40' \
		-H "Content-type: application/octet-stream" \
		"http://localhost:3500/v1.0/publish/messages/topic1"

.PHONY: invoke
invoke: ## Invokes service method with different operations
	curl -i -d '{ "from": "John", "to": "Lary", "message": "hi" }' \
		-H "Content-type: application/json" \
		"http://localhost:3500/v1.0/invoke/serving/method/echo"
	curl -i -d "ping" \
		-H "Content-type: text/plain;charset=UTF-8" \
		"http://localhost:3500/v1.0/invoke/serving/method/echo"
	curl -i -X DELETE \
		"http://localhost:3500/v1.0/invoke/serving/method/echo?k1=v1&k2=v2"

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
