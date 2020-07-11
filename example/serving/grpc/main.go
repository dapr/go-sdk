package main

import (
	"context"
	"log"

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

func echoHandler(ctx context.Context, typeIn string, dataIn []byte) (typeOut string, dataOut []byte) {
	content := string(dataIn)
	log.Printf("content: %s", content)
	return "text/plain; charset=UTF-8", []byte(content)
}

func testHandler(ctx context.Context, typeIn string, dataIn []byte) (typeOut string, dataOut []byte) {
	return "text/plain; charset=UTF-8", []byte("tessting")
}

// Topic Subscriptions

func messageHandler(ctx context.Context, e *daprd.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}
