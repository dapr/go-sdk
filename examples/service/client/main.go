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
	"time"

	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	// just for this demo
	ctx := context.Background()
	json := `{ "message": "hello" }`
	data := []byte(json)
	store := "statestore"
	pubsub := "messages"

	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// publish a message to the topic demo
	if err := client.PublishEvent(ctx, pubsub, "demo", data); err != nil {
		panic(err)
	}
	fmt.Println("data published")

	// save state with the key key1
	fmt.Printf("saving data: %s\n", string(data))
	if err := client.SaveState(ctx, store, "key1", data, nil); err != nil {
		panic(err)
	}
	fmt.Println("data saved")

	// get state for key key1
	item, err := client.GetState(ctx, store, "key1", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("data retrieved [key:%s etag:%s]: %s\n", item.Key, item.Etag, string(item.Value))

	// save state with options
	item2 := &dapr.SetStateItem{
		Etag: &dapr.ETag{
			Value: "2",
		},
		Key: item.Key,
		Metadata: map[string]string{
			"created-on": time.Now().UTC().String(),
		},
		Value: item.Value,
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyLastWrite,
			Consistency: dapr.StateConsistencyStrong,
		},
	}
	if err := client.SaveBulkState(ctx, store, item2); err != nil {
		panic(err)
	}
	fmt.Println("data item saved")

	// delete state for key key1
	if err := client.DeleteState(ctx, store, "key1", nil); err != nil {
		panic(err)
	}
	fmt.Println("data deleted")

	// invoke a method called EchoMethod on another dapr enabled service
	content := &dapr.DataContent{
		ContentType: "text/plain",
		Data:        []byte("hellow"),
	}
	resp, err := client.InvokeMethodWithContent(ctx, "serving", "echo", "post", content)
	if err != nil {
		panic(err)
	}
	fmt.Printf("service method invoked, response: %s\n", string(resp))

	in := &dapr.InvokeBindingRequest{
		Name:      "example-http-binding",
		Operation: "create",
	}
	if err := client.InvokeOutputBinding(ctx, in); err != nil {
		panic(err)
	}
	fmt.Println("output binding invoked")

	fmt.Println("DONE (CTRL+C to Exit)")
}
