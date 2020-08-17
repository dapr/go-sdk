# Dapr SDK for Go

Client library to accelerate Dapr application development in go. This client supports all public [Dapr API](https://github.com/dapr/docs/tree/master/reference/api) and focuses on developer productivity. 

[![Test](https://github.com/dapr/go-sdk/workflows/Test/badge.svg)](https://github.com/dapr/go-sdk/actions?query=workflow%3ATest) [![Release](https://github.com/dapr/go-sdk/workflows/Release/badge.svg)](https://github.com/dapr/go-sdk/actions?query=workflow%3ARelease) [![Go Report Card](https://goreportcard.com/badge/github.com/dapr/go-sdk)](https://goreportcard.com/report/github.com/dapr/go-sdk) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/dapr/go-sdk)

## Usage

> Assuming you already have [installed](https://golang.org/doc/install) go

Dapr go client includes two packages: `client` (for invoking public Dapr API) and `service` (to create services in go that can be invoked by Dapr, this is sometimes refereed to as "callback"). 

### Client 

Import Dapr go `client` package:

```go
import "github.com/dapr/go-sdk/client"
```

#### Quick start

```go
package main

import (
    dapr "github.com/dapr/go-sdk/client"
)

func main() {
    client, err := dapr.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()
    //TODO: use the client here, see below for examples 
}
```

Assuming you have [Dapr CLI](https://github.com/dapr/docs/blob/master/getting-started/environment-setup.md) installed locally, you can then launch your app locally like this:

```shell
dapr run --app-id example-service \
         --app-protocol grpc \
         --app-port 50001 \
         go run main.go
```

Check the [example folder](./example) for working Dapr go client examples.

To accelerate your Dapr service development even more, consider the GitHub templates with complete gRPC solutions for two common use-cases:

* [gRPC Event Subscriber Template](https://github.com/mchmarny/dapr-grpc-event-subscriber-template) for pub/sub event processing 
* [gRPC Serving Service Template ](https://github.com/mchmarny/dapr-grpc-service-template) which creates a target for service to service invocations 


#### Usage

The Dapr go client supports following functionality: 

##### State 

For simple use-cases, Dapr client provides easy to use methods for `Save`, `Get`, and `Delete`: 

```go
ctx := context.Background()
data := []byte("hello")
store := "my-store" // defined in the component YAML 

// save state with the key key1
if err := client.SaveState(ctx, store, "key1", data); err != nil {
    panic(err)
}

// get state for key key1
item, err := client.GetState(ctx, store, "key1")
if err != nil {
    panic(err)
}
fmt.Printf("data [key:%s etag:%s]: %s", item.Key, item.Etag, string(item.Value))

// delete state for key key1
if err := client.DeleteState(ctx, store, "key1"); err != nil {
    panic(err)
}
```

For more granular control, the Dapr go client exposed `SetStateItem` type which can be use to gain more control over the state operations and allow for multiple items to be saved at once:

```go     
item1 := &dapr.SetStateItem{
    Key:  "key1",
    Etag: "2",
    Metadata: map[string]string{
        "created-on": time.Now().UTC().String(),
    },
    Value: []byte("hello"),
    Options: &dapr.StateOptions{
        Concurrency: dapr.StateConcurrencyLastWrite,
        Consistency: dapr.StateConsistencyStrong,
    },
}
item2 := &dapr.SetStateItem{
    Key:  "key2",
    Etag: "1",
    Metadata: map[string]string{
        "created-on": time.Now().UTC().String(),
    },
    Value: []byte("hello again"),
    Options: &dapr.StateOptions{
        Concurrency: dapr.StateConcurrencyLastWrite,
        Consistency: dapr.StateConsistencyStrong,
    },
}
if err := client.SaveStateItems(ctx, store, item1, item2); err != nil {
    panic(err)
}
```

Similarly, `GetStateItems` method provides a way to retrieve multiple state items in a single operation:

```go
keys := []string{"key1", "key2", "key3"}
items, err := GetStateItems(ctx, store, keys, parallelism 100)
```

And the `ExecuteStateTransaction` method to transactionally execute multiple `upsert` or `delete` operations.

```go
ops := make([]*dapr.StateOperation, 0)

op1 := &dapr.StateOperation{
    Type: dapr.StateOperationTypeUpsert,
    Item: &dapr.SetStateItem{
        Key:   "key1",
        Value: []byte(data),
    },
}
op2 := &dapr.StateOperation{
    Type: dapr.StateOperationTypeDelete,
    Item: &dapr.SetStateItem{
        Key:   "key2",
    },
}
ops = append(ops, op1, op2)
meta := map[string]string{}
err := testClient.ExecuteStateTransaction(ctx, store, meta, ops)
```

##### PubSub 

To publish data onto a topic the Dapr client provides a simple method:

```go
data := []byte(`{ "id": "a123", "value": "abcdefg", "valid": true }`)
err = client.PublishEvent(ctx, "topic-name", data)
handleErrors(err)
```

##### Service Invocation 

To invoke a specific method on another service running with Dapr sidecar, the Dapr client provides two options. To invoke a service without any data:

```go 
resp, err = client.InvokeService(ctx, "service-name", "method-name") 
handleErrors(err)
``` 

And to invoke a service with data: 

```go 
data := []byte(`{ "id": "a123", "value": "abcdefg", "valid": true }`)
resp, err := client.InvokeServiceWithContent(ctx, "service-name", "method-name", "application/json", data)
handleErrors(err)
```

##### Bindings

Similarly to Service, Dapr client provides two methods to invoke an operation on a [Dapr-defined binding](https://github.com/dapr/docs/tree/master/concepts/bindings). Dapr supports input, output, and bidirectional bindings so the first methods supports all of them along with metadata options: 

```go
data := []byte("hello")
opt := map[string]string{
    "opt1": "value1",
    "opt2": "value2",
}
resp, meta, err := client.InvokeBinding(ctx, "binding-name", "operation-name", data, opt)
handleErrors(err)
```

And for simple, output only biding:

```go
data := []byte("hello")
err = client.InvokeOutputBinding(ctx, "binding-name", "operation-name", data)
handleErrors(err)
```

##### Secrets

The Dapr client also provides access to the runtime secrets that can be backed by any number of secrete stores (e.g. Kubernetes Secrets, Hashicorp Vault, or Azure KeyVault):

```go
opt := map[string]string{
    "version": "2",
}
secret, err = client.GetSecret(ctx, "store-name", "secret-name", opt)
handleErrors(err)
```

## Service (callback)

In addition to a an easy to use client, Dapr go package also provides implementation for `service`. Instructions on how to use it are located [here](./service/Readme.md)


## Contributing to Dapr go client 

See the [Contribution Guide](./CONTRIBUTING.md) to get started with building and developing.
