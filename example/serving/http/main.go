package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	daprd "github.com/dapr/go-sdk/service/http"
)

func main() {
	// create a regular HTTP server mux
	mux := http.NewServeMux()

	// create a Dapr service
	s, err := daprd.NewService(mux)
	if err != nil {
		log.Fatalf("error creating sever: %v", err)
	}

	// add some topic subscriptions
	err = s.AddTopicEventHandler("messages", "/messages", messageHandler)
	if err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// handle all the added topic handlers
	err = s.HandleSubscriptions()
	if err != nil {
		log.Fatalf("error creating topic subscription: %v", err)
	}

	invokeHandler := func(ctx context.Context, in *daprd.InvocationEvent) (out []byte, err error) {
		if in == nil {
			err = errors.New("nil invocation parameter")
			return
		}
		log.Printf("echo handler (%s): %+v", in.ContentType, string(in.Data))
		out = in.Data
		return
	}

	err = s.AddInvocationHandler("/EchoMethod", invokeHandler)
	if err != nil {
		log.Fatalf("error adding invocation handler: %v", err)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func messageHandler(ctx context.Context, e daprd.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}
