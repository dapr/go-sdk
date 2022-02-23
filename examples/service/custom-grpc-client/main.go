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

	"google.golang.org/grpc"

	dapr "github.com/dapr/go-sdk/client"
)

func GetEnvValue(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Testing 40 MB data exchange
	maxRequestBodySize := 40
	var opts []grpc.CallOption

	// Receive 40 MB + 1 MB (data + headers overhead) exchange
	headerBuffer := 1
	opts = append(opts, grpc.MaxCallRecvMsgSize((maxRequestBodySize+headerBuffer)*1024*1024))
	conn, err := grpc.Dial(net.JoinHostPort("127.0.0.1",
		GetEnvValue("DAPR_GRPC_PORT", "50001")),
		grpc.WithDefaultCallOptions(opts...), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	// Instantiate DAPR client with custom-grpc-client gRPC connection
	client := dapr.NewClientWithConnection(conn)
	defer client.Close()

	ctx := context.Background()
	start := time.Now()

	fmt.Println("Writing large data blob...")
	data := make([]byte, maxRequestBodySize*1024*1024)
	store := "statestore" // defined in the component YAML
	key := "my_key"

	// save state with the key my_key, default options: strong, last-write
	if err := client.SaveState(ctx, store, key, data, nil); err != nil {
		panic(err)
	}
	fmt.Println("Saved the large data blob...")
	elapsed := time.Since(start)
	fmt.Printf("Writing to statestore took %s", elapsed)

	// get state for key my_key
	fmt.Println("Getting data from the large data blob...")
	_, err = client.GetState(ctx, store, key, nil)
	if err != nil {
		panic(err)
	}
	elapsed2 := time.Since(start)
	fmt.Printf("Reading from statestore took %s\n", elapsed2)

	// delete state for key my_key
	if err := client.DeleteState(ctx, store, key, nil); err != nil {
		panic(err)
	}
	elapsed3 := time.Since(start)
	fmt.Printf("Deleting key from statestore took %s\n", elapsed3)

	fmt.Println("DONE (CTRL+C to Exit)")
}
