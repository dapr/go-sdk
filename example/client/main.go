package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/dapr/go-sdk/pkg/dapr"
)

// server is our user app
type server struct {
}

func main() {
	dapr.AddInvokeHandler(`MyMethod`, MyMethod)
	dapr.AddBindingHandler(`storage`, storage)
	dapr.AddTopicHandler(`TopicA`, TopicA)

	// and start...
	if err := dapr.Serve(`:4000`); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// MyMethod is a sample method to invoke
func MyMethod(ctx context.Context, args proto.Message) (result proto.Message, err error) {
	return &wrappers.StringValue{Value: `Hi there!`}, nil
}

func storage(ctx context.Context, args proto.Message) (result proto.Message, err error) {
	fmt.Println("Invoked from binding")
	return &empty.Empty{}, nil
}

// TopicA is a sample topic handler
func TopicA(ctx context.Context, data proto.Message) error {
	fmt.Println("Topic message arrived")
	return nil
}
