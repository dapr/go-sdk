# Dapr Service (Callback) SDK for Go

In addition to this Dapr API client, Dapr Go SDK also provides `service` package to bootstrap your Dapr callback services. These services can be developed in either gRPC or HTTP:

* [HTTP Service](./http/Readme.md)
* [gRPC Service](./grpc/Readme.md)

## Templates 

To accelerate your Dapr app development in Go even further, we've created a few GitHub templates which build on the above Dapr callback packages:

* [Dapr gRPC Service in Go](https://github.com/dapr-templates/dapr-grpc-service-template) - Template project to jump start your Dapr event subscriber service with gRPC development
* [Dapr HTTP Event Subscriber in Go](https://github.com/dapr-templates/dapr-http-event-subscriber-template) - Template project to jump start your Dapr event subscriber service with HTTP development
* [Dapr gRPC Event Subscriber in Go](https://github.com/dapr-templates/dapr-grpc-event-subscriber-template) - Template project to jump start your Dapr event subscriber service with gRPC development
* [Dapr HTTP cron Handler in Go](https://github.com/dapr-templates/dapr-http-cron-handler-template) - Template project to jump start your Dapr service development for scheduled workloads

## Contributing

See the [Contribution Guide](../CONTRIBUTING.md) to get started with building and developing.
