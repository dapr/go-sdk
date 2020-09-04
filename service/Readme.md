# Dapr Service (Callback) SDK for Go

In addition to this Dapr API client, Dapr go SDK also provides `service` package to bootstrap your Dapr callback services. These services can be developed in either gRPC or HTTP:

* [HTTP Service](./http/Readme.md)
* [gRPC Service](./grpc/Readme.md)

## Templates 

To accelerate your Dapr app development in go even further, we've craated a few GitHub templates which build on the above Dapr callback packages:

* [dapr-grpc-event-subscriber-template](https://github.com/dapr/dapr-grpc-event-subscriber-template)
* [dapr-grpc-service-template](https://github.com/dapr/dapr-grpc-service-template)
* [dapr-http-event-subscriber-template](https://github.com/dapr/dapr-http-event-subscriber-template)
* [dapr-http-cron-handler-template](https://github.com/dapr/dapr-http-cron-handler-template)

## Contributing

See the [Contribution Guide](../CONTRIBUTING.md) to get started with building and developing.
