package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	stateStoreName = `statestore`
	socket         = "/tmp/dapr-order-app-grpc.socket"
)

func main() {
	var orderID int
	put := kingpin.Command("put", "Send a new order.")
	put.Flag("id", "order ID.").Default("1").IntVar(&orderID)
	kingpin.Command("get", "Get current order.")
	kingpin.Command("del", "Delete the order.")
	kingpin.Command("seq", "Stream sequence of orders.")

	// create the client
	client, err := dapr.NewClientWithSocket(socket)
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
