package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"

	pbc "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/daprclient/v1"

	"google.golang.org/grpc"
)

// server is our user app
type server struct {
}

func main() {
	// create listener
	lis, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create gRPC server
	s := grpc.NewServer()
	pb.RegisterDaprClientServer(s, &server{})

	fmt.Println("client starting...")

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Sample method to invoke
func (s *server) MyMethod() string {
	return "pong"
}

// This method gets invoked when a remote service has called the app through dapr
// The payload carries a Method to identify the method, a set of metadata properties and an optional payload
func (s *server) OnInvoke(ctx context.Context, in *pbc.InvokeRequest) (*pbc.InvokeResponse, error) {
	var response string

	fmt.Println(fmt.Sprintf("got invoked with: %s", string(in.Data.Value)))

	switch in.Method {
	case "MyMethod":
		response = s.MyMethod()
	}

	return &pbc.InvokeResponse{
		ContentType: "text/plain; charset=UTF-8",
		Data:        &any.Any{Value: []byte(response)},
	}, nil
}

// GetTopicSubscriptions will call this method to get the list of topics the app wants to subscribe to.
// In this example, we are telling dapr. To subscribe to a topic named example-topic
func (s *server) GetTopicSubscriptions(ctx context.Context, in *empty.Empty) (*pb.GetTopicSubscriptionsEnvelope, error) {
	return &pb.GetTopicSubscriptionsEnvelope{
		Topics: []string{"example-topic"},
	}, nil
}

// GetBindingsSubscriptions will call this method to get the list of bindings the app will get invoked by. In this example, we are telling dapr
// To invoke our app with a binding named storage
func (s *server) GetBindingsSubscriptions(ctx context.Context, in *empty.Empty) (*pb.GetBindingsSubscriptionsEnvelope, error) {
	return &pb.GetBindingsSubscriptionsEnvelope{
		Bindings: []string{"example-storage"},
	}, nil
}

// OnBindingEvent method gets invoked every time a new event is fired from a registered binding.
// The message carries the binding name, a payload and optional metadata
func (s *server) OnBindingEvent(ctx context.Context, in *pb.BindingEventEnvelope) (*pb.BindingResponseEnvelope, error) {
	fmt.Println("Invoked from binding")
	return &pb.BindingResponseEnvelope{}, nil
}

// OnTopicEvent method is fired whenever a message has been published to a topic that has been subscribed.
// dapr sends published messages in a CloudEvents 0.3 envelope.
func (s *server) OnTopicEvent(ctx context.Context, in *pb.CloudEventEnvelope) (*empty.Empty, error) {
	fmt.Println("Topic message arrived")
	return &empty.Empty{}, nil
}
