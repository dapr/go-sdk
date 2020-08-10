# Dapr SDK for Go

This is the dapr SDK (client) for go (golang). It covers all of the APIs described in Dapr's [protocol buffers](https://raw.githubusercontent.com/dapr/dapr/master/dapr/proto/) with focus on developer productivity. 

[![Test](https://github.com/dapr/go-sdk/workflows/Test/badge.svg)](https://github.com/dapr/go-sdk/actions?query=workflow%3ATest) [![Release](https://github.com/dapr/go-sdk/workflows/Release/badge.svg)](https://github.com/dapr/go-sdk/actions?query=workflow%3ARelease) [![Go Report Card](https://goreportcard.com/badge/github.com/dapr/go-sdk)](https://goreportcard.com/report/github.com/dapr/go-sdk) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/dapr/go-sdk)

## Installation

To install Dapr client package, you need to first [install go](https://golang.org/doc/install) and set up your development environment. To import Dapr go client in your code:

```go
import "github.com/dapr/go-sdk/client"
```

## Quick start

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
    //TODO: use the client here 
}
```

Assuming you have Dapr CLI installed locally, you can then launch your app like this:

```shell
dapr run --app-id my-app --app-protocol grpc --app-port 50001 go run main.go
```

See [example folder](./example) for complete example. 


## Examples

Few common Dapr client usage examples 

### State 

For simple use-cases, Dapr client provides easy to use methods: 

```go
ctx := context.Background()
data := []byte("hello")
store := "my-store" // defined in the component YAML 

// save state with the key
err = client.SaveStateData(ctx, store, "k1", "v1", data)
handleErrors(err)

// get state for key
out, etag, err := client.GetState(ctx, store, "k1")
handleErrors(err)

// delete state for key
err = client.DeleteState(ctx, store, "k1")
handleErrors(err)
```

The `StateItem` type exposed by Dapr client provides more granular control options:

```go     
data := &client.StateItem{
    Etag:     "v1",
    Key:      "k1",
    Metadata: map[string]string{
        "key1": "value1",
        "key2": "value2",
    },
    Value:    []byte("hello"),
    Options:  &client.StateOptions{
        Concurrency: client.StateConcurrencyLastWrite,
        Consistency: client.StateConsistencyStrong,
        RetryPolicy: &client.StateRetryPolicy{
            Threshold: 3,
            Pattern: client.RetryPatternExponential,
            Interval: time.Duration(5 * time.Second),
        },
    },
}
err = client.SaveStateItem(ctx, store, data)
```

Similar `StateOptions` exist on `GetDate` and `DeleteState` methods. Additionally, Dapr client also provides a method to save multiple state items at once:

```go 
data := &client.State{
    StoreName: "my-store",
    States: []*client.StateItem{
        {
            Key:   "k1",
            Value: []byte("message 1"),
        },
        {
            Key:   "k2",
            Value: []byte("message 2"),
        },
    },
}
err = client.SaveState(ctx, data)
```

### PubSub 

To publish data onto a topic the Dapr client provides a simple method:

```go
data := []byte("hello")
err = client.PublishEvent(ctx, "topic-name", data)
handleErrors(err)
```

### Service Invocation 

To invoke a specific method on another service running with Dapr sidecar, the Dapr client provides two options. To invoke a service without any data:

```go 
resp, err = client.InvokeService(ctx, "service-name", "method-name") 
handleErrors(err)
``` 

And to invoke a service with data: 

```go 
data := []byte("hello")
resp, err := client.InvokeServiceWithContent(ctx, "service-name", "method-name", "text/plain", data)
handleErrors(err)
```

### Bindings

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

### Secrets

The Dapr client also provides access to the runtime secrets that can be backed by any number of secrete stores (e.g. Kubernetes Secrets, Hashicorp Vault, or Azure KeyVault):

```go
opt := map[string]string{
    "version": "2",
}
secret, err = client.GetSecret(ctx, "store-name", "secret-name", opt)
handleErrors(err)
```

## Contributing to Dapr go client 

See the [Contribution Guide](./CONTRIBUTING.md) to get started with building and developing.
