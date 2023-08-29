package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dapr/go-sdk/client"
)

func main() {
	// initialise dapr client
	dapr, err := client.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// obtain metadata
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	metadata, err := dapr.GetMetadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// print metadata
	fmt.Printf("Sidecar id: %s\n", metadata.Id)

	fmt.Printf("Sidecar extended metadata that should not exist: %s\n", metadata.ExtendedMetadata["test_string"])

	// set test metadata
	ctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	err = dapr.SetMetadata(ctx, "test_string", "test_string_exists")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Set test metadata")

	ctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	metadata, err = dapr.GetMetadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sidecar extended metadata that should exist: %s\n", metadata.ExtendedMetadata["test_string"])

	fmt.Println(metadata.ExtendedMetadata)
}
