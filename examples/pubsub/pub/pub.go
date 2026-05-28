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
	"fmt"
	"net"
	"os"
	"time"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	// set the environment as instructions.
	pubsubName = os.Getenv("DAPR_PUBSUB_NAME")
	topicName  = "neworder"
)

func main() {
	ctx := context.Background()
	publishEventData := []byte("ping")
	publishEventsData := []interface{}{"multi-ping", "multi-pong"}

	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Wait for the subscriber to be ready before publishing
	for range 30 {
		if conn, err := net.DialTimeout("tcp", "localhost:8080", time.Second); err == nil {
			conn.Close()
			break
		}
		time.Sleep(time.Second)
	}

	// Publish a single event
	if err := client.PublishEvent(ctx, pubsubName, topicName, publishEventData); err != nil {
		panic(err)
	}

	// Publish multiple events
	if res := client.PublishEvents(ctx, pubsubName, topicName, publishEventsData); res.Error != nil {
		panic(res.Error)
	}

	fmt.Println("data published")

	fmt.Println("Done (CTRL+C to Exit)")
}
