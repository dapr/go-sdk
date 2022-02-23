---
type: docs
title: "Getting started with the Dapr client Go SDK"
linkTitle: "Client"
weight: 20000
description: How to get up and running with the Dapr Go SDK
no_list: true
---

The Dapr client package allows you to interact with other Dapr applications from a Go application.

## Prerequisites

- [Dapr CLI]({{< ref install-dapr-cli.md >}}) installed
- Initialized [Dapr environment]({{< ref install-dapr-selfhost.md >}})
- [Go installed](https://golang.org/doc/install)


## Import the client package 
```go
import "github.com/dapr/go-sdk/client"
```

## Building blocks

The Go SDK allows you to interface with all of the [Dapr building blocks]({{< ref building-blocks >}}).

### Service Invocation

To invoke a specific method on another service running with Dapr sidecar, the Dapr client Go SDK provides two options:

Invoke a service without data:
```go
resp, err := client.InvokeMethod(ctx, "app-id", "method-name", "post")
```

Invoke a service with data:
```go
content := &dapr.DataContent{
    ContentType: "application/json",
    Data:        []byte(`{ "id": "a123", "value": "demo", "valid": true }`),
}

resp, err = client.InvokeMethodWithContent(ctx, "app-id", "method-name", "post", content)
```

- For a full guide on service invocation visit [How-To: Invoke a service]({{< ref howto-invoke-discover-services.md >}}).

### State Management

For simple use-cases, Dapr client provides easy to use `Save`, `Get`, `Delete` methods:

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

For more granular control, the Dapr Go client exposes `SetStateItem` type, which can be use to gain more control over the state operations and allow for multiple items to be saved at once:

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
items, err := client.GetBulkState(ctx, store, keys, nil,100)
```

And the `ExecuteStateTransaction` method to execute multiple upsert or delete operations transactionally.

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

### Publish Messages
To publish data onto a topic, the Dapr Go client provides a simple method:

```go
data := []byte(`{ "id": "a123", "value": "abcdefg", "valid": true }`)
if err := client.PublishEvent(ctx, "component-name", "topic-name", data); err != nil {
    panic(err)
}
```

- For a full list of state operations visit [How-To: Publish & subscribe]({{< ref howto-publish-subscribe.md >}}).

### Output Bindings
The Dapr Go client SDK provides two methods to invoke an operation on a Dapr-defined binding. Dapr supports input, output, and bidirectional bindings.

For simple, output only biding:
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


- For a full guide on output bindings visit [How-To: Use bindings]({{< ref howto-bindings.md >}}).

### Secret Management

The Dapr client also provides access to the runtime secrets that can be backed by any number of secrete stores (e.g. Kubernetes Secrets, HashiCorp Vault, or Azure KeyVault):

```go
opt := map[string]string{
    "version": "2",
}

secret, err := client.GetSecret(ctx, "store-name", "secret-name", opt)
```

### Authentication

By default, Dapr relies on the network boundary to limit access to its API. If however the target Dapr API is configured with token-based authentication, users can configure the Go Dapr client with that token in two ways:

**Environment Variable**

If the DAPR_API_TOKEN environment variable is defined, Dapr will automatically use it to augment its Dapr API invocations to ensure authentication.

**Explicit Method**

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


- For a full guide on secrets visit [How-To: Retrieve secrets]({{< ref howto-secrets.md >}}).

## Related links
- [Go SDK Examples](https://github.com/dapr/go-sdk/tree/main/examples)
