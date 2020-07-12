package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	daprd "github.com/dapr/go-sdk/service/http"
)

func main() {
	// create a Dapr service
	s := daprd.NewService()

	// add some topic subscriptions
	err := s.AddTopicEventHandler("messages", "/messages", messageHandler)
	if err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// add a service to service invocation handler
	err = s.AddInvocationHandler("/EchoMethod", echoHandler)
	if err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}

	// start service on address (e.g. ":8080", "0.0.0.0:8080", "10.1.1.1:8080" )
	if err = s.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func echoHandler(ctx context.Context, in *daprd.InvocationEvent) (out []byte, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}
	log.Printf("echo handler (%s): %+v", in.ContentType, string(in.Data))
	out = in.Data
	return
}

func messageHandler(ctx context.Context, e daprd.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}
