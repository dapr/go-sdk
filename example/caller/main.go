package main

import (
	"context"
	"fmt"

	"github.com/dapr/go-sdk/pkg/dapr"
	"github.com/golang/protobuf/ptypes/wrappers"
)

func main() {

	// Create the client
	client, err := dapr.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	// Invoke a method called MyMethod on another Dapr enabled service with id client
	req := &wrappers.StringValue{Value: `Hello`}
	resp := &wrappers.StringValue{}
	err = client.Invoke(context.Background(), `client`, `MyMethod`, req, resp)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Value)
	}

	// Publish a message to the topic TopicA
	message := &wrappers.StringValue{Value: `Hi from Pub Sub`}
	err = client.Publish(context.Background(), `TopicA`, message)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Published message!")
	}

	// Save state with the key myKey
	err = client.SaveState(context.Background(), &dapr.State{
		Key:   `myKey`,
		Value: &wrappers.StringValue{Value: `My State`},
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Saved state!")
	}

	// Get state for key myKey
	r := &wrappers.StringValue{}
	err = client.GetState(context.Background(), `myKey`, r)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Got state!")
		fmt.Println(r.Value)
	}

	// Delete state for key myKey
	err = client.DeleteState(context.Background(), `myKey`)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("State deleted")
	}

	// Invoke output binding named storage. Make sure you set up a Dapr binding, otherwise this will fail
	data := &wrappers.StringValue{Value: `some data`}
	err = client.Binding(context.Background(), `storage`, data)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Binding invoked")
	}
}
