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
## Error handling
Dapr errors are based on [gRPC's richer error model](https://cloud.google.com/apis/design/errors#error_model). 
The following code shows an example of how you can parse and handle the error details:

```go
if err != nil {
    st := status.Convert(err)

    fmt.Printf("Code: %s\n", st.Code().String())
    fmt.Printf("Message: %s\n", st.Message())

    for _, detail := range st.Details() {
        switch t := detail.(type) {
        case *errdetails.ErrorInfo:
            // Handle ErrorInfo details
            fmt.Printf("ErrorInfo:\n- Domain: %s\n- Reason: %s\n- Metadata: %v\n", t.GetDomain(), t.GetReason(), t.GetMetadata())
        case *errdetails.BadRequest:
            // Handle BadRequest details
            fmt.Println("BadRequest:")
            for _, violation := range t.GetFieldViolations() {
                fmt.Printf("- Key: %s\n", violation.GetField())
                fmt.Printf("- The %q field was wrong: %s\n", violation.GetField(), violation.GetDescription())
            }
        case *errdetails.ResourceInfo:
            // Handle ResourceInfo details
            fmt.Printf("ResourceInfo:\n- Resource type: %s\n- Resource name: %s\n- Owner: %s\n- Description: %s\n",
                t.GetResourceType(), t.GetResourceName(), t.GetOwner(), t.GetDescription())
        case *errdetails.Help:
            // Handle ResourceInfo details
            fmt.Println("HelpInfo:")
            for _, link := range t.GetLinks() {
                fmt.Printf("- Url: %s\n", link.Url)
                fmt.Printf("- Description: %s\n", link.Description)
            }
        
        default:
            // Add cases for other types of details you expect
            fmt.Printf("Unhandled error detail type: %v\n", t)
        }
    }
}
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

For a full guide on service invocation, visit [How-To: Invoke a service]({{< ref howto-invoke-discover-services.md >}}).

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

Retrieve, filter, and sort key/value data stored in your statestore using `QueryState`. 

```go
// Define the query string
query := `{
	"filter": {
		"EQ": { "value.Id": "1" }
	},
	"sort": [
		{
			"key": "value.Balance",
			"order": "DESC"
		}
	]
}`

// Use the client to query the state
queryResponse, err := c.QueryState(ctx, "querystore", query)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Got %d\n", len(queryResponse))

for _, account := range queryResponse {
	var data Account
	err := account.Unmarshal(&data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account: %s has %f\n", data.ID, data.Balance)
}
```

> **Note:** Query state API is currently in alpha

For a full guide on state management, visit [How-To: Save & get state]({{< ref howto-get-save-state.md >}}).

### Publish Messages
To publish data onto a topic, the Dapr Go client provides a simple method:

```go
data := []byte(`{ "id": "a123", "value": "abcdefg", "valid": true }`)
if err := client.PublishEvent(ctx, "component-name", "topic-name", data); err != nil {
    panic(err)
}
```

To publish multiple messages at once, the `PublishEvents` method can be used:

```go
events := []string{"event1", "event2", "event3"}
res := client.PublishEvents(ctx, "component-name", "topic-name", events)
if res.Error != nil {
    panic(res.Error)
}
```

For a full guide on pub/sub, visit [How-To: Publish & subscribe]({{< ref howto-publish-subscribe.md >}}).

### Output Bindings


The Dapr Go client SDK provides two methods to invoke an operation on a Dapr-defined binding. Dapr supports input, output, and bidirectional bindings.

For simple, output-only binding:

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

For a full guide on output bindings, visit [How-To: Use bindings]({{< ref howto-bindings.md >}}).

### Actors

Use the Dapr Go client SDK to write actors.

```go
// MyActor represents an example actor type.
type MyActor struct {
	actors.Actor
}

// MyActorMethod is a method that can be invoked on MyActor.
func (a *MyActor) MyActorMethod(ctx context.Context, req *actors.Message) (string, error) {
	log.Printf("Received message: %s", req.Data)
	return "Hello from MyActor!", nil
}

func main() {
	// Create a Dapr client
	daprClient, err := client.NewClient()
	if err != nil {
		log.Fatal("Error creating Dapr client: ", err)
	}

	// Register the actor type with Dapr
	actors.RegisterActor(&MyActor{})

	// Create an actor client
	actorClient := actors.NewClient(daprClient)

	// Create an actor ID
	actorID := actors.NewActorID("myactor")

	// Get or create the actor
	err = actorClient.SaveActorState(context.Background(), "myactorstore", actorID, map[string]interface{}{"data": "initial state"})
	if err != nil {
		log.Fatal("Error saving actor state: ", err)
	}

	// Invoke a method on the actor
	resp, err := actorClient.InvokeActorMethod(context.Background(), "myactorstore", actorID, "MyActorMethod", &actors.Message{Data: []byte("Hello from client!")})
	if err != nil {
		log.Fatal("Error invoking actor method: ", err)
	}

	log.Printf("Response from actor: %s", resp.Data)

	// Wait for a few seconds before terminating
	time.Sleep(5 * time.Second)

	// Delete the actor
	err = actorClient.DeleteActor(context.Background(), "myactorstore", actorID)
	if err != nil {
		log.Fatal("Error deleting actor: ", err)
	}

	// Close the Dapr client
	daprClient.Close()
}
```

For a full guide on actors, visit [the Actors building block documentation]({{< ref actors >}}).

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


For a full guide on secrets, visit [How-To: Retrieve secrets]({{< ref howto-secrets.md >}}).

### Distributed Lock

The Dapr client provides mutually exclusive access to a resource using a lock. With a lock, you can:

- Provide access to a database row, table, or an entire database
- Lock reading messages from a queue in a sequential manner

```go
package main

import (
    "fmt"

    dapr "github.com/dapr/go-sdk/client"
)

func main() {
    client, err := dapr.NewClient()
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    resp, err := client.TryLockAlpha1(ctx, "lockstore", &dapr.LockRequest{
			LockOwner:         "random_id_abc123",
			ResourceID:      "my_file_name",
			ExpiryInSeconds: 60,
		})

    fmt.Println(resp.Success)
}
```

For a full guide on distributed lock, visit [How-To: Use a lock]({{< ref howto-use-distributed-lock.md >}}).

### Configuration

With the Dapr client Go SDK, you can consume configuration items that are returned as read-only key/value pairs, and subscribe to configuration item changes.

#### Config Get

```go
	items, err := client.GetConfigurationItem(ctx, "example-config", "mykey")
	if err != nil {
		panic(err)
	}
	fmt.Printf("get config = %s\n", (*items).Value)
```

#### Config Subscribe

```go
go func() {
	if err := client.SubscribeConfigurationItems(ctx, "example-config", []string{"mySubscribeKey1", "mySubscribeKey2", "mySubscribeKey3"}, func(id string, items map[string]*dapr.ConfigurationItem) {
		for k, v := range items {
			fmt.Printf("get updated config key = %s, value = %s \n", k, v.Value)
		}
		subscribeID = id
	}); err != nil {
		panic(err)
	}
}()
```

For a full guide on configuration, visit [How-To: Manage configuration from a store]({{< ref howto-manage-configuration.md >}}).

### Cryptography

With the Dapr client Go SDK, you can use the high-level `Encrypt` and `Decrypt` cryptography APIs to encrypt and decrypt files while working on a stream of data.

To encrypt:

```go
// Encrypt the data using Dapr
out, err := client.Encrypt(context.Background(), rf, dapr.EncryptOptions{
	// These are the 3 required parameters
	ComponentName: "mycryptocomponent",
	KeyName:        "mykey",
	Algorithm:     "RSA",
})
if err != nil {
	panic(err)
}
```

To decrypt:

```go
// Decrypt the data using Dapr
out, err := client.Decrypt(context.Background(), rf, dapr.EncryptOptions{
	// Only required option is the component name
	ComponentName: "mycryptocomponent",
})
```

For a full guide on cryptography, visit [How-To: Use the cryptography APIs]({{< ref howto-cryptography.md >}}).

## Related links
[Go SDK Examples](https://github.com/dapr/go-sdk/tree/main/examples)
