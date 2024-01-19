package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"

	"github.com/dapr/kit/logger"
)

var log = logger.NewLogger("dapr.examples.grpc-client")

const (
	address = "localhost:50007"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "dapr-app-id", "grpc-server")
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "Dapr"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Infof("Greeting: %s", r.GetMessage())
}
