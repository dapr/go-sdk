<<<<<<< HEAD
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
	err = client.SaveStateWithData(ctx, "statestore", "key1", data)
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
=======
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"

	commonv1pb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"google.golang.org/grpc"
)

// server is our user app
type server struct {
}

func main() {
	// create listiner
	lis, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	s := grpc.NewServer()
	pb.RegisterAppCallbackServer(s, &server{})

	fmt.Println("Client starting...")

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Sample method to invoke
func (s *server) MyMethod() string {
	return "Hi there!"
}

// This method gets invoked when a remote service has called the app through Dapr
// The payload carries a Method to identify the method, a set of metadata properties and an optional payload
func (s *server) OnInvoke(ctx context.Context, in *commonv1pb.InvokeRequest) (*commonv1pb.InvokeResponse, error) {
	var response string

	fmt.Println(fmt.Sprintf("Got invoked with: %s", string(in.Data.Value)))

	switch in.Method {
	case "MyMethod":
		response = s.MyMethod()
	}

	return &commonv1pb.InvokeResponse{
		ContentType: "text/plain; charset=UTF-8",
		Data:        &any.Any{Value: []byte(response)},
	}, nil
}

// Dapr will call this method to get the list of topics the app wants to subscribe to. In this example, we are telling Dapr
// To subscribe to a topic named TopicA
func (s *server) ListTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.ListTopicSubscriptionsResponse, error) {
	return &pb.ListTopicSubscriptionsResponse{
		Subscriptions: []*pb.TopicSubscription{
			{Topic: "TopicA"},
		},
	}, nil
}

// Dapr will call this method to get the list of bindings the app will get invoked by. In this example, we are telling Dapr
// To invoke our app with a binding named storage
func (s *server) ListInputBindings(ctx context.Context, in *empty.Empty) (*pb.ListInputBindingsResponse, error) {
	return &pb.ListInputBindingsResponse{
		Bindings: []string{"storage"},
	}, nil
}

// This method gets invoked every time a new event is fired from a registerd binding. The message carries the binding name, a payload and optional metadata
func (s *server) OnBindingEvent(ctx context.Context, in *pb.BindingEventRequest) (*pb.BindingEventResponse, error) {
	fmt.Println("Invoked from binding")
	return &pb.BindingEventResponse{}, nil
}

// This method is fired whenever a message has been published to a topic that has been subscribed. Dapr sends published messages in a CloudEvents 0.3 envelope.
func (s *server) OnTopicEvent(ctx context.Context, in *pb.TopicEventRequest) (*empty.Empty, error) {
	fmt.Println("Topic message arrived")
	return &empty.Empty{}, nil
}
>>>>>>> upstream/master
