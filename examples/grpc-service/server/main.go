package main

import (
	"context"
	"net"

	daprd "github.com/dapr/go-sdk/service/grpc"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"github.com/dapr/kit/logger"
)

var log = logger.NewLogger("dapr.examples.grpc-server")

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Infof("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	daprServer := daprd.NewServiceWithGrpcServer(lis, s)

	// start the server
	if err := daprServer.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
