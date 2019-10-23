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
func MyMethod(ctx context.Context, args proto.Message, meta map[string]string) (result proto.Message, err error) {
	return &wrappers.StringValue{Value: `Hi there!`}, nil
}

func storage(ctx context.Context, args proto.Message, meta map[string]string) (result proto.Message, options []dapr.Option, err error) {
	fmt.Println("Invoked from binding")
	return &empty.Empty{}, nil, nil
}

// TopicA is a sample topic handler
func TopicA(ctx context.Context, data proto.Message) error {
	fmt.Println("Topic message arrived")
	return nil
}
