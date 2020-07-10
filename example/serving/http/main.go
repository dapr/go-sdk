package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dapr/go-sdk/server/event"
	daprd "github.com/dapr/go-sdk/server/http"
)

func main() {
	// create a regular HTTP server mux
	mux := http.NewServeMux()

	// create a Dapr service server
	daprServer, err := daprd.NewServer(mux)
	if err != nil {
		log.Fatalf("error creating sever: %v", err)
	}

	// add some topic subscriptions
	err = daprServer.AddTopicEventHandler("messages", "/messages", messageHandler)
	if err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	// start the server
	err = daprServer.HandleSubscriptions()
	if err != nil {
		log.Fatalf("error creating topic subscription: %v", err)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func messageHandler(ctx context.Context, e event.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %s", e.Topic, e.ID, string(e.Data))
	return nil
}
