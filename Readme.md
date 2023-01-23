# Dapr SDK for Go

Client library to help you build Dapr application in Go. This client supports all public [Dapr APIs](https://docs.dapr.io/reference/api/) while focusing on idiomatic Go experience and developer productivity. 

[![Test](https://github.com/dapr/go-sdk/workflows/Test/badge.svg)](https://github.com/dapr/go-sdk/actions?query=workflow%3ATest) [![Release](https://github.com/dapr/go-sdk/workflows/Release/badge.svg)](https://github.com/dapr/go-sdk/actions?query=workflow%3ARelease) [![Go Report Card](https://goreportcard.com/badge/github.com/dapr/go-sdk)](https://goreportcard.com/report/github.com/dapr/go-sdk) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/dapr/go-sdk) [![codecov](https://codecov.io/gh/dapr/go-sdk/branch/main/graph/badge.svg)](https://codecov.io/gh/dapr/go-sdk) [![FOSSA Status](https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fdapr%2Fgo-sdk.svg?type=shield)](https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fdapr%2Fgo-sdk?ref=badge_shield)

## Usage
> Assuming you already have [installed](https://golang.org/doc/install) Go

Dapr Go client includes two packages: `client` (for invoking public Dapr APIs), and `service` (to create services that will be invoked by Dapr, this is sometimes referred to as "callback").

### Creating client 

Import Dapr Go `client` package:

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
    // TODO: use the client here, see below for examples 
}
```

`NewClient` function has a default timeout for 5s, but you can customize this timeout by setting the environment variable `DAPR_CLIENT_TIMEOUT_SECONDS`.  
For example: 
```go
package main

import (
	"os"
	
	dapr "github.com/dapr/go-sdk/client"
)

func main() {
    os.Setenv("DAPR_CLIENT_TIMEOUT_SECONDS", "3")
    client, err := dapr.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()
}
```
  
Assuming you have [Dapr CLI](https://docs.dapr.io/getting-started/install-dapr-cli/) installed, you can then launch your app locally like this:

```shell
dapr run --app-id example-service \
         --app-protocol grpc \
         --app-port 50001 \
         go run main.go
```

See the [example folder](./examples) for more working Dapr client examples.

#### Usage

The Go client supports all the building blocks exposed by Dapr API. Let's review these one by one: 


##### State 

For simple use-cases, Dapr client provides easy to use `Save`, `Get`, and `Delete` methods: 

```go
ctx := context.Background()
data := []byte("hello")
store := "my-store" // defined in the component YAML 

// save state with the key key1, default options: strong, last-write
if err := client.SaveState(ctx, store, "key1", data, nil); err != nil {
    panic(err)
}

// get state for key key1
item, err := client.GetState(ctx, store, "key1", nil)
if err != nil {
    panic(err)
}
fmt.Printf("data [key:%s etag:%s]: %s", item.Key, item.Etag, string(item.Value))

// delete state for key key1
if err := client.DeleteState(ctx, store, "key1", nil); err != nil {
    panic(err)
}
```

For more granular control, the Dapr Go client exposes `SetStateItem` type, which can be used to gain more control over the state operations and allow for multiple items to be saved at once:

```go     
item1 := &dapr.SetStateItem{
    Key:  "key1",
    Etag: &ETag{
        Value: "1",
    },
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
    Etag: &dapr.ETag{
	Value: "1",
    },
    Value: []byte("hello again"),
}

if err := client.SaveBulkState(ctx, store, item1, item2, item3); err != nil {
    panic(err)
}
```

Similarly, `GetBulkState` method provides a way to retrieve multiple state items in a single operation:

```go
keys := []string{"key1", "key2", "key3"}
items, err := client.GetBulkState(ctx, store, keys, nil, 100)
```

And the `ExecuteStateTransaction` method to execute multiple `upsert` or `delete` operations transactionally.

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

To publish data onto a topic, the Dapr client provides a simple method:

```go
data := []byte(`{ "id": "a123", "value": "abcdefg", "valid": true }`)
if err := client.PublishEvent(ctx, "component-name", "topic-name", data); err != nil {
    panic(err)
}
```

##### Service Invocation 

To invoke a specific method on another service running with Dapr sidecar, the Dapr client provides two options. To invoke a service without any data:

```go 
resp, err := client.InvokeMethod(ctx, "app-id", "method-name", "post")
``` 

And to invoke a service with data: 

```go 
content := &dapr.DataContent{
    ContentType: "application/json",
    Data:        []byte(`{ "id": "a123", "value": "demo", "valid": true }`),
}

resp, err = client.InvokeMethodWithContent(ctx, "app-id", "method-name", "post", content)
```

##### Bindings

Similarly to Service, Dapr client provides two methods to invoke an operation on a [Dapr-defined binding](https://docs.dapr.io/developing-applications/building-blocks/bindings/). Dapr supports input, output, and bidirectional bindings.

For simple, output only binding:

```go
in := &dapr.InvokeBindingRequest{ Name: "binding-name", Operation: "operation-name" }
err = client.InvokeOutputBinding(ctx, in)
```

To invoke method with content and metadata:

```go
in := &dapr.InvokeBindingRequest{
    Name:      "binding-name",
    Operation: "operation-name",
    Data: []byte("hello"),
    Metadata: map[string]string{"k1": "v1", "k2": "v2"},
}

out, err := client.InvokeBinding(ctx, in)
```

##### Secrets

The Dapr client also provides access to the runtime secrets that can be backed by any number of secret stores (e.g. Kubernetes Secrets, HashiCorp Vault, or Azure KeyVault):

```go
opt := map[string]string{
    "version": "2",
}

secret, err := client.GetSecret(ctx, "store-name", "secret-name", opt)
```

##### Distributed Lock 

The Dapr client provides methods to grab a distributed lock and unlock it.

Grab a lock:

```go
ctx := context.Background()
store := "my-store" // defined in the component YAML 

r, err := testClient.TryLockAlpha1(ctx, testLockStore, &LockRequest{
    LockOwner:         "owner1",
	ResourceID:      "resource1",
    ExpiryInSeconds: 5,
})
```

Unlock a lock:

```go
r, err := testClient.UnlockAlpha1(ctx, testLockStore, &UnlockRequest{
	LockOwner:    "owner1",
	ResourceID: "resource1",
})
```

##### Authentication

By default, Dapr relies on the network boundary to limit access to its API. If however the target Dapr API is configured with token-based authentication, users can configure the Go Dapr client with that token in two ways:

###### Environment Variable

If the `DAPR_API_TOKEN` environment variable is defined, Dapr will automatically use it to augment its Dapr API invocations to ensure authentication. 

###### Explicit Method

In addition, users can also set the API token explicitly on any Dapr client instance. This approach is helpful in cases when the user code needs to create multiple clients for different Dapr API endpoints. 

```go
func main() {
    client, err := dapr.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()
    client.WithAuthToken("your-Dapr-API-token-here")
}
```

### Service (callback)

In addition to the client capabilities that allow you to call into the Dapr API, the Go SDK also provides `service` package to help you bootstrap Dapr callback services in either gRPC or HTTP. Instructions on how to use it are located [here](./service/Readme.md).

## Contributing to Dapr Go client

See the [Contribution Guide](./CONTRIBUTING.md) to get started with building and developing.

## Code of Conduct

Please refer to our [Dapr Community Code of Conduct](https://github.com/dapr/community/blob/master/CODE-OF-CONDUCT.md).