# Dapr Protos

These are generated protos from the [Dapr](https://github.com/dapr/dapr/tree/master/dapr) repository.

They are version locked using the [.dapr-proto-ref](../../.dapr-proto-ref) file.

### Bumping the version of the proto files

The command `make proto-update` makes it easier to bump the proto files version.

### Generating protos

To generate the protos, using the current locked version, use the command `make proto`
