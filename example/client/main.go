package main

import (
	"context"
	"fmt"
	"time"

	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	// just for this demo
	ctx := context.Background()
	data := []byte("ping")
	store := "statestore"

	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// publish a message to the topic messagebus
	if err := client.PublishEvent(ctx, "messagebus", "my-topic", data); err != nil {
		panic(err)
	}
	fmt.Println("data published")

	// save state with the key key1
	if err := client.SaveState(ctx, store, "key1", data); err != nil {
		panic(err)
	}
	fmt.Println("data saved")

	// get state for key key1
	item, err := client.GetState(ctx, store, "key1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("data [key:%s etag:%s]: %s", item.Key, item.Etag, string(item.Value))

	// save state with options
	item2 := &dapr.SetStateItem{
		Etag: "2",
		Key:  item.Key,
		Metadata: map[string]string{
			"created-on": time.Now().UTC().String(),
		},
		Value: item.Value,
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyLastWrite,
			Consistency: dapr.StateConsistencyStrong,
		},
	}
	if err := client.SaveStateItems(ctx, store, item2); err != nil {
		panic(err)
	}

	// delete state for key key1
	if err := client.DeleteState(ctx, store, "key1"); err != nil {
		panic(err)
	}
	fmt.Println("data deleted")

	// invoke a method called EchoMethod on another dapr enabled service
	resp, err := client.InvokeServiceWithContent(ctx, "serving", "echo", "text/plain", data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("service method invoked, response: %s", string(resp))

	if err := client.InvokeOutputBinding(ctx, "example-http-binding", "create", nil); err != nil {
		panic(err)
	}
	fmt.Println("output binding invoked")
	fmt.Println("DONE (CTRL+C to Exit)")
}
