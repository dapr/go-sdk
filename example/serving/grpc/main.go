package main

import (
	"context"
	"errors"
	"log"

	"github.com/dapr/go-sdk/service"
	daprd "github.com/dapr/go-sdk/service/grpc"
)

func main() {
	// create a Dapr service server
	s, err := daprd.NewService(":50001")
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	// add some topic subscriptions
	err = s.AddTopicEventHandler("messages", eventHandler)
	if err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// add a service to service invocation handler
	err = s.AddServiceInvocationHandler("EchoMethod", echoHandler)
	if err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}

	// add a binding invocation handler
	err = s.AddBindingInvocationHandler("run", runHandler)
	if err != nil {
		log.Fatalf("error adding binding handler: %v", err)
	}

	// start the server
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func echoHandler(ctx context.Context, in *service.InvocationEvent) (out *service.InvocationEvent, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}
	log.Printf("echo handler (%s): %+v", in.ContentType, string(in.Data))
	out = in
	return
}

func eventHandler(ctx context.Context, e *service.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}

func runHandler(ctx context.Context, in *service.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%v, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
