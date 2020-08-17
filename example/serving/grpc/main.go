package main

import (
	"context"
	"errors"
	"log"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/grpc"
)

func main() {
	// create a Dapr service server
	s, err := daprd.NewService(":50001")
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	// add some topic subscriptions
<<<<<<< HEAD
	sub := &common.Subscription{
		Topic: "messages",
	}
	if err := s.AddTopicEventHandler(sub, eventHandler); err != nil {
=======
	err = s.AddTopicEventHandler("messages", "demo", eventHandler)
	if err != nil {
>>>>>>> mchmarny-multi-pubsub-support-clean
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// add a service to service invocation handler
	if err := s.AddServiceInvocationHandler("echo", echoHandler); err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}

	// add a binding invocation handler
	if err := s.AddBindingInvocationHandler("run", runHandler); err != nil {
		log.Fatalf("error adding binding handler: %v", err)
	}

	// start the server
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

<<<<<<< HEAD
func eventHandler(ctx context.Context, e *common.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
=======
func eventHandler(ctx context.Context, e *daprd.TopicEvent) error {
	log.Printf("event - PubsubName:%s, Topic:%s, ID:%s, Data: %v", e.PubsubName, e.Topic, e.ID, e.Data)
>>>>>>> mchmarny-multi-pubsub-support-clean
	return nil
}

func echoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}
	log.Printf(
		"echo - ContentType:%s, Verb:%s, QueryString:%s, %+v",
		in.ContentType, in.Verb, in.QueryString, string(in.Data),
	)
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func runHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%v, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
