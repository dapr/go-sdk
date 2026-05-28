/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"log"
	"os"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	// set the environment as instructions.
	pubsubName = "messages"
	topicName1 = "sendorder"
	topicName2 = "neworder"
	logger     = log.New(os.Stdout, "", log.LstdFlags)
)

func main() {
	ctx := context.Background()
	publishEventData := []byte("ping123")
	publishEventsData := []interface{}{"multi-ping", "multi-pong"}

	client, err := dapr.NewClient()
	if err != nil {
		logger.Fatalf("error creating dapr client: %v", err)
	}
	defer client.Close()

	// Publish a single event
	logger.Println("sending message")
	if err := client.PublishEvent(ctx, pubsubName, topicName1, publishEventData); err != nil {
		logger.Fatalf("error publishing event: %v", err)
	}
	if err := client.PublishEvent(ctx, pubsubName, topicName2, publishEventData); err != nil {
		logger.Fatalf("error publishing event: %v", err)
	}
	logger.Println("message published")

	// Publish multiple events
	logger.Println("sending multiple messages")
	if res := client.PublishEvents(ctx, pubsubName, topicName1, publishEventsData); res.Error != nil {
		logger.Fatalf("error publishing events: %v", res.Error)
	}
	if res := client.PublishEvents(ctx, pubsubName, topicName2, publishEventsData); res.Error != nil {
		logger.Fatalf("error publishing events: %v", res.Error)
	}
	logger.Println("multiple messages published")
}
