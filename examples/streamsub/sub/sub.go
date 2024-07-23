/*
Copyright 2024 The Dapr Authors
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
	"errors"
	"fmt"
	"log"
	"time"

	daprd "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
)

func main() {
	client, err := daprd.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	deadLetterTopic := "deadletter"

	// Streaming subscription for topic "sendorder" on pubsub component
	// "messages". The given subscription handler is called when a message is
	// received. The  returned `stop` function is used to stop the subscription
	// and close the connection.
	stop, err := client.SubscribeWithHandler(context.Background(),
		daprd.SubscriptionOptions{
			PubsubName:      "messages",
			Topic:           "sendorder",
			DeadLetterTopic: &deadLetterTopic,
		},
		eventHandler,
	)
	if err != nil {
		log.Fatalf("failed to subscribe to topic: %v", err)
	}
	fmt.Printf(">>Created subscription messages/sendorder\n")

	// Another method of streaming subscriptions, this time for the topic "neworder".
	// The returned `sub` object is used to receive messages.
	// `sub` must be closed once it's no longer needed.
	sub, err := client.Subscribe(context.Background(), daprd.SubscriptionOptions{
		PubsubName:      "messages",
		Topic:           "neworder",
		DeadLetterTopic: &deadLetterTopic,
	})
	if err != nil {
		log.Fatalf("failed to subscribe to topic: %v", err)
	}
	fmt.Printf(">>Created subscription messages/neworder\n")

	for i := 0; i < 3; i++ {
		msg, err := sub.Receive()
		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}
		log.Printf(">>Received message\n")
		log.Printf("event - PubsubName: %s, Topic: %s, ID: %s, Data: %s\n", msg.PubsubName, msg.Topic, msg.ID, msg.RawData)

		// Use _MUST_ always signal the result of processing the message, else the
		// message will not be considered as processed and will be redelivered or
		// dead lettered.
		if err := msg.Success(); err != nil {
			log.Fatalf("error sending message success: %v", err)
		}
	}

	time.Sleep(time.Second * 10)

	if err := errors.Join(stop(), sub.Close()); err != nil {
		log.Fatal(err)
	}
}

func eventHandler(e *common.TopicEvent) common.SubscriptionResponseStatus {
	log.Printf(">>Received message\n")
	log.Printf("event - PubsubName: %s, Topic: %s, ID: %s, Data: %s\n", e.PubsubName, e.Topic, e.ID, e.Data)
	return common.SubscriptionResponseStatusSuccess
}
