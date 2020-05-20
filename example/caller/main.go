package main

import (
	"context"
	"fmt"
	"os"

	commonv1pb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
)

func main() {
	// Get the Dapr port and create a connection
	daprPort := os.Getenv("DAPR_GRPC_PORT")
	daprAddress := fmt.Sprintf("localhost:%s", daprPort)
	conn, err := grpc.Dial(daprAddress, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Create the client
	client := pb.NewDaprClient(conn)

	// Invoke a method called MyMethod on another Dapr enabled service with id client
	resp, err := client.InvokeService(context.Background(), &pb.InvokeServiceRequest{
		Id: "client",
		Message: &commonv1pb.InvokeRequest{
			Method:      "MyMethod",
			ContentType: "text/plain; charset=UTF-8",
			Data:        &any.Any{Value: []byte("Hello")},
		},
	})
	if err != nil {
		panic(err)
	}

	if resp.GetContentType() != "text/plain; charset=UTF-8" {
		fmt.Printf("wrong content type: %s", resp.GetContentType())
	}

	fmt.Println(string(resp.GetData().GetValue()))

	// Publish a message to the topic TopicA
	_, err = client.PublishEvent(context.Background(), &pb.PublishEventRequest{
		Topic: "TopicA",
		Data:  []byte("Hi from Pub Sub"),
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Published message!")
	}

	// Save state with the key myKey
	_, err = client.SaveState(context.Background(), &pb.SaveStateRequest{
		// statestore is the name of the default redis state store , set up by Dapr CLI
		StoreName: "statestore",
		Requests: []*commonv1pb.StateSaveRequest{
			{
				Key:   "myKey",
				Value: []byte("My State"),
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Saved state!")
	}

	// Get state for key myKey
	r, err := client.GetState(context.Background(), &pb.GetStateRequest{
		// statestore is the name of the default redis state store , set up by Dapr CLI
		StoreName: "statestore",
		Key:       "myKey",
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Got state!")
		fmt.Println(string(r.Data))
	}

	// Delete state for key myKey
	_, err = client.DeleteState(context.Background(), &pb.DeleteStateRequest{
		// statestore is the name of the default redis state store , set up by Dapr CLI
		StoreName: "statestore",
		Key:       "myKey",
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("State deleted")
	}

	// Invoke output binding named storage. Make sure you set up a Dapr binding, otherwise this will fail
	_, err = client.InvokeBinding(context.Background(), &pb.InvokeBindingRequest{
		Name: "storage",
		Data: []byte("some data"),
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Binding invoked")
	}
}
