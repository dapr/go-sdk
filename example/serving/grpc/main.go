package main

import (
	"context"
	"log"

	daprd "github.com/dapr/go-sdk/service/grpc"
)

func main() {
	// create a Dapr service server
	server, err := daprd.NewService(":50001")
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	// add some invocation handlers
	server.AddInvocationHandler("EchoMethod", echoHandler)

	// add some topic subscriptions
	server.AddTopicEventHandler("messages", messageHandler)

	// start the server
	if err := server.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// Invocation Handlers

func echoHandler(ctx context.Context, in *daprd.InvocationEvent) (out *daprd.InvocationEvent, err error) {
	log.Printf("content: %+v", in)
	out = &daprd.InvocationEvent{
		ContentType: in.ContentType,
		Data:        in.Data,
	}
	return
}

// Topic Subscriptions

func messageHandler(ctx context.Context, e *daprd.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}
