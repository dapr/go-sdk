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
	resp, err := client.InvokeService(ctx, "my-client", "MyMethod", data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))

	// publish a message to the topic my-topic
	err = client.PublishEvent(ctx, "my-topic", data)
	if err != nil {
		panic(err)
	}
	fmt.Println("data published")

	// save state with the key key1
	err = client.SaveState(ctx, "my-store", "key1", data)
	if err != nil {
		panic(err)
	}
	fmt.Println("data saved")

	// get state for key key1
	dataOut, err := client.GetState(ctx, "my-store", "key1")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dataOut))

	// delete state for key key1
	err = client.DeleteState(ctx, "my-store", "key1")
	if err != nil {
		panic(err)
	}
	fmt.Println("data deleted")

	// invoke output binding named 'kafka-topic-name'.
	// make sure you set up a dapr binding, otherwise this will fail
	err = client.InvokeBinding(ctx, "kafka-topic-name", data)
	if err != nil {
		panic(err)
	}
	fmt.Println("binding invoked")
}
