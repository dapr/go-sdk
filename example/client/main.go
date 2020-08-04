package main

import (
	"context"
	"log"
	"os"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

func main() {
	// just for this demo
	ctx := context.Background()
	data := []byte("ping")

	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		logger.Panic(err)
	}
	defer client.Close()

	// publish a message to the topic messagebus
	err = client.PublishEvent(ctx, "messagebus", data)
	if err != nil {
		logger.Panic(err)
	}
	logger.Println("data published")

	// save state with the key key1
	err = client.SaveStateData(ctx, "statestore", "key1", "1", data)
	if err != nil {
		logger.Panic(err)
	}
	logger.Println("data saved")

	// get state for key key1
	dataOut, etag, err := client.GetState(ctx, "statestore", "key1")
	if err != nil {
		logger.Panic(err)
	}
	logger.Printf("data out [etag:%s]: %s", etag, string(dataOut))

	// delete state for key key1
	err = client.DeleteState(ctx, "statestore", "key1")
	if err != nil {
		logger.Panic(err)
	}
	logger.Println("data deleted")

	// invoke a method called EchoMethod on another dapr enabled service
	resp, err := client.InvokeServiceWithContent(ctx, "serving", "EchoMethod",
		"text/plain; charset=UTF-8", data)
	if err != nil {
		logger.Panic(err)
	}
	logger.Printf("service method invoked, response: %s", string(resp))

	err = client.InvokeOutputBinding(ctx, "example-http-binding", "create", nil)
	if err != nil {
		panic(err)
	}
	logger.Println("binding invoked")
}
