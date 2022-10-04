package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/metadata"
)

func init() {
	opts := &redis.Options{
		Addr: "127.0.0.1:6379",
	}
	client := redis.NewClient(opts)
	// set config value
	client.Set(context.Background(), "mykey", "myConfigValue", -1)
	ticker := time.NewTicker(time.Second)
	go func() {
		for i := 0; i < 5; i++ {
			<-ticker.C
			// update config value
			client.Set(context.Background(), "mySubscribeKey1", "mySubscribeValue"+strconv.Itoa(i+1), -1)
			client.Set(context.Background(), "mySubscribeKey2", "mySubscribeValue"+strconv.Itoa(i+1), -1)
			client.Set(context.Background(), "mySubscribeKey3", "mySubscribeValue"+strconv.Itoa(i+1), -1)
		}
		ticker.Stop()
	}()
}

func main() {
	ctx := context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}

	items, err := client.GetConfigurationItem(ctx, "example-config", "mykey")
	if err != nil {
		panic(err)
	}
	fmt.Printf("get config = %s\n", (*items).Value)

	ctx, f := context.WithTimeout(ctx, 60*time.Second)
	md := metadata.Pairs("dapr-app-id", "configuration-api")
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer f()
	var subscribeID string
	go func() {
		if err := client.SubscribeConfigurationItems(ctx, "example-config", []string{"mySubscribeKey1", "mySubscribeKey2", "mySubscribeKey3"}, func(id string, items map[string]*dapr.ConfigurationItem) {
			for k, v := range items {
				fmt.Printf("get updated config key = %s, value = %s \n", k, v.Value)
			}
			subscribeID = id
		}); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second*3 + time.Millisecond*500)

	// dapr configuration unsubscribe called.
	if err := client.UnsubscribeConfigurationItems(ctx, "example-config", subscribeID); err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 5)
}
