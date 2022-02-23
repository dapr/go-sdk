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
	"os"
	"strconv"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	dapr "github.com/dapr/go-sdk/client"
)

const (
	stateStoreName = `statestore`
	daprPort       = "3500"
)

var port string

func init() {
	if port = os.Getenv("DAPR_GRPC_PORT"); len(port) == 0 {
		port = daprPort
	}
}

func main() {
	var orderID int
	put := kingpin.Command("put", "Send a new order.")
	put.Flag("id", "order ID.").Default("1").IntVar(&orderID)
	kingpin.Command("get", "Get current order.")
	kingpin.Command("del", "Delete the order.")
	kingpin.Command("seq", "Stream sequence of orders.")

	// create the client
	client, err := dapr.NewClientWithPort(port)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	ctx := context.Background()

	switch kingpin.Parse() {

	case "get":
		fmt.Printf("Getting order\n")
		item, err := client.GetState(ctx, stateStoreName, "order", nil)
		if err != nil {
			fmt.Printf("Failed to get state: %v\n", err)
		}
		if len(item.Value) > 0 {
			fmt.Printf("Order ID %s\n", item.Value)
		} else {
			fmt.Printf("Order Not Found\n")
		}
	case "put":
		fmt.Printf("Sending order ID %d\n", orderID)
		err := client.SaveState(ctx, stateStoreName, "order", []byte(strconv.Itoa(orderID)), nil)
		if err != nil {
			fmt.Printf("Failed to persist state: %v\n", err)
		} else {
			fmt.Printf("Successfully persisted state\n")
		}
	case "del":
		fmt.Printf("Deleting order\n")
		err := client.DeleteState(ctx, stateStoreName, "order", nil)
		if err != nil {
			fmt.Printf("Failed to delete state: %v\n", err)
		} else {
			fmt.Printf("Successfully deleted state\n")
		}
	case "seq":
		fmt.Printf("Streaming sequence of orders\n")
		for {
			orderID++
			err := client.SaveState(ctx, stateStoreName, "order", []byte(strconv.Itoa(orderID)), nil)
			if err != nil {
				fmt.Printf("Failed to persist state: %v\n", err)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}
