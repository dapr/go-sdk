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
	"net/http"

	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
)

// Subscription to tell the dapr what topic to subscribe.
// - PubsubName: is the name of the component configured in the metadata of pubsub.yaml.
// - Topic: is the name of the topic to subscribe.
// - Route: tell dapr where to request the API to publish the message to the subscriber when get a message from topic.
// - Match: (Optional) The CEL expression to match on the CloudEvent to select this route.
// - Priority: (Optional) The priority order of the route when Match is specificed.
//             If not specified, the matches are evaluated in the order in which they are added.
var defaultSubscription = &common.Subscription{
	PubsubName: "messages",
	Topic:      "neworder",
	Route:      "/orders",
}

var importantSubscription = &common.Subscription{
	PubsubName: "messages",
	Topic:      "neworder",
	Route:      "/important",
	Match:      `event.type == "important"`,
	Priority:   1,
}

func main() {
	s := daprd.NewService(":8080")

	if err := s.AddTopicEventHandler(defaultSubscription, eventHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	if err := s.AddTopicEventHandler(importantSubscription, importantEventHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func eventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	log.Printf("event - PubsubName: %s, Topic: %s, ID: %s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)
	return false, nil
}

func importantEventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	log.Printf("important event - PubsubName: %s, Topic: %s, ID: %s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)
	return false, nil
}
