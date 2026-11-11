package main

import (
	"context"
	"log"
	"net"
	"os"

	daprd "github.com/dapr/go-sdk/service/grpc"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50051"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	logger.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	daprServer := daprd.NewServiceWithGrpcServer(lis, s)

	// start the server
	if err := daprServer.Start(); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
