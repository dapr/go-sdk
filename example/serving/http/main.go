package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
)

func main() {
	// create a Dapr service (e.g. ":8080", "0.0.0.0:8080", "10.1.1.1:8080" )
	s := daprd.NewService(":8080")

	// add some topic subscriptions
	sub := &common.Subscription{
		PubsubName: "messages",
		Topic:      "topic1",
		Route:      "/events",
	}
	if err := s.AddTopicEventHandler(sub, eventHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// add a service to service invocation handler
	if err := s.AddServiceInvocationHandler("/echo", echoHandler); err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}

	// add an input binding invocation handler
	if err := s.AddBindingInvocationHandler("/run", runHandler); err != nil {
		log.Fatalf("error adding binding handler: %v", err)
	}

	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func eventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	log.Printf("event - PubsubName:%s, Topic:%s, ID:%s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)
	return false, nil
}

func echoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("invocation parameter required")
		return
	}
	log.Printf(
		"echo - ContentType:%s, Verb:%s, QueryString:%s, %s",
		in.ContentType, in.Verb, in.QueryString, in.Data,
	)
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func runHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%s, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
