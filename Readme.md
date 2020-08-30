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
    Metadata: map[string]string{
        "created-on": time.Now().UTC().String(),
    },
    Value: []byte("hello again"),
}

item3 := &dapr.SetStateItem{
    Key:  "key3",
    Etag: "1",
    Value: []byte("hello again"),
}

if err := client.SaveStateItems(ctx, store, item1, item2, item3); err != nil {
    panic(err)
}
```

Similarly, `GetBulkItems` method provides a way to retrieve multiple state items in a single operation:

```go
keys := []string{"key1", "key2", "key3"}
items, err := client.GetBulkItems(ctx, store, keys, 100)
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
if err := client.PublishEvent(ctx, "component-name", "topic-name", data); err != nil {
    panic(err)
}
```

##### Service Invocation 

To invoke a specific method on another service running with Dapr sidecar, the Dapr client provides two options. To invoke a service without any data:

```go 
resp, err = client.InvokeService(ctx, "service-name", "method-name") 
``` 

And to invoke a service with data: 

```go 
content := &dapr.DataContent{
    ContentType: "application/json",
    Data:        []byte(`{ "id": "a123", "value": "demo", "valid": true }`),
}

resp, err := client.InvokeServiceWithContent(ctx, "service-name", "method-name", content)
```

##### Bindings

Similarly to Service, Dapr client provides two methods to invoke an operation on a [Dapr-defined binding](https://github.com/dapr/docs/tree/master/concepts/bindings). Dapr supports input, output, and bidirectional bindings.

For simple, output only biding:

```go
in := &dapr.BindingInvocation{ Name: "binding-name", Operation: "operation-name" }
err = client.InvokeOutputBinding(ctx, in)
```

To invoke method with content and metadata:

```go
in := &dapr.BindingInvocation{
    Name:      "binding-name",
    Operation: "operation-name",
    Data: []byte("hello"),
    Metadata: map[string]string{"k1": "v1", "k2": "v2"},
}

out, err := client.InvokeBinding(ctx, in)
```

##### Secrets

The Dapr client also provides access to the runtime secrets that can be backed by any number of secrete stores (e.g. Kubernetes Secrets, Hashicorp Vault, or Azure KeyVault):

```go
opt := map[string]string{
    "version": "2",
}

secret, err := client.GetSecret(ctx, "store-name", "secret-name", opt)
```

## Service (callback)

In addition to this Dapr API client, Dapr go SDK also provides `service` package to bootstrap your Dapr callback services in either gRPC or HTTP. Instructions on how to use it are located [here](./service/Readme.md)

## Contributing to Dapr go client 

See the [Contribution Guide](./CONTRIBUTING.md) to get started with building and developing.
