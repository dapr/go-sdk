package main

import (
	"context"
	"fmt"

	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	ctx := context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}

	items, err := client.ConverseAlpha1(ctx, "testllm", &dapr.ConversationRequest{
		Inputs: []dapr.ConversationInput{
			{
				Message: "hi there",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(items.Outputs[0].Result)
}
