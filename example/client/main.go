package main

import (
	"context"
	"fmt"

	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	// just for this demo
	ctx := context.Background()
	data := []byte("ping")

	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close(ctx)

	// invoke a method called MyMethod on another dapr enabled service with id client
	resp, err := client.InvokeService(ctx, "serving", "MyMethod", data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))

	// publish a message to the topic example-topic
	err = client.PublishEvent(ctx, "example-topic", data)
	if err != nil {
		panic(err)
	}
	fmt.Println("data published")

	// save state with the key key1
	err = client.SaveState(ctx, "example-store", "key1", data)
	if err != nil {
		panic(err)
	}
	fmt.Println("data saved")

	// get state for key key1
	dataOut, err := client.GetState(ctx, "example-store", "key1")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dataOut))

	// delete state for key key1
	err = client.DeleteState(ctx, "example-store", "key1")
	if err != nil {
		panic(err)
	}
	fmt.Println("data deleted")

	// invoke output binding named 'kafka-topic'.
	// make sure you set up a dapr binding, otherwise this will fail
	err = client.InvokeBinding(ctx, "kafka-topic", data)
	if err != nil {
		panic(err)
	}
	fmt.Println("binding invoked")
}
