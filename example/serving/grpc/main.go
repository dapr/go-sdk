package main

import (
	"context"
	"log"

	"github.com/dapr/go-sdk/server/event"
	daprd "github.com/dapr/go-sdk/server/grpc"
)

func main() {
	// create a Dapr service server
	server, err := daprd.NewServer("50001")
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	// add some invocation handlers
	server.AddInvocationHandler("EchoMethod", echoHandler)
	server.AddInvocationHandler("Test", testHandler)

	// add some topic subscriptions
	server.AddTopicEventHandler("messages", messageHandler)

	// start the server
	if err := server.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// Invocation Handlers

func echoHandler(ctx context.Context, contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte) {
	content := string(dataIn)
	log.Printf("content: %s", content)
	return "text/plain; charset=UTF-8", []byte(content)
}

func testHandler(ctx context.Context, contentTypeIn string, dataIn []byte) (contentTypeOut string, dataOut []byte) {
	return "text/plain; charset=UTF-8", []byte("tessting")
}

// Topic Subscriptions

func messageHandler(ctx context.Context, event *event.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %s", event.Topic, event.ID, string(event.Data))
	return nil
}
