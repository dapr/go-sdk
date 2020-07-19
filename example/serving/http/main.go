package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	daprd "github.com/dapr/go-sdk/service/http"
)

func main() {
	// create a Dapr service (e.g. ":8080", "0.0.0.0:8080", "10.1.1.1:8080" )
	s := daprd.NewService(":8080")

	// add some topic subscriptions
	err := s.AddTopicEventHandler("messages", "/events", eventHandler)
	if err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// add a service to service invocation handler
	err = s.AddServiceInvocationHandler("/echo", echoHandler)
	if err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}

	// add a binding invocation handler
	err = s.AddBindingInvocationHandler("/run", runHandler)
	if err != nil {
		log.Fatalf("error adding binding handler: %v", err)
	}

	if err = s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func eventHandler(ctx context.Context, e *daprd.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}

func echoHandler(ctx context.Context, in *daprd.InvocationEvent) (out *daprd.InvocationEvent, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}
	log.Printf("echo handler (%s): %+v", in.ContentType, string(in.Data))
	out = in
	return
}

func runHandler(ctx context.Context, in *daprd.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%v, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
