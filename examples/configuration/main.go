package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/metadata"
)

func addItems(wg *sync.WaitGroup) {
	opts := &redis.Options{
		Addr: "127.0.0.1:6379",
	}
	client := redis.NewClient(opts)
	// set config value
	client.Set(context.Background(), "mykey", "myConfigValue", -1)
	ticker := time.NewTicker(time.Second)
	wg.Add(3 * 5)
	go func() {
		for i := 0; i < 5; i++ {
			<-ticker.C
			client.Set(context.Background(), "mySubscribeKey1", "mySubscribeValue"+strconv.Itoa(i+1), -1)
			client.Set(context.Background(), "mySubscribeKey2", "mySubscribeValue"+strconv.Itoa(i+1), -1)
			client.Set(context.Background(), "mySubscribeKey3", "mySubscribeValue"+strconv.Itoa(i+1), -1)
		}
		ticker.Stop()
	}()
}

func main() {
	var wg sync.WaitGroup
	addItems(&wg)
	ctx := context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}

	items, err := client.GetConfigurationItem(ctx, "example-config", "mykey")
	if err != nil {
		panic(err)
	}
	fmt.Printf("got config key = mykey, value = %s \n", (*items).Value)

	ctx, f := context.WithTimeout(ctx, 60*time.Second)
	md := metadata.Pairs("dapr-app-id", "configuration-api")
	ctx = metadata.NewOutgoingContext(ctx, md)
	defer f()
	subscribeID, err := client.SubscribeConfigurationItems(ctx, "example-config", []string{"mySubscribeKey1", "mySubscribeKey2", "mySubscribeKey3"}, func(id string, items map[string]*dapr.ConfigurationItem) {
		wg.Done()
		for k, v := range items {
			fmt.Printf("got config key = %s, value = %s \n", k, v.Value)
		}
	})
	if err != nil {
		panic(err)
	}
	wg.Wait()

	// dapr configuration unsubscribe called.
	if err := client.UnsubscribeConfigurationItems(ctx, "example-config", subscribeID); err != nil {
		panic(err)
	}
	fmt.Println("dapr configuration unsubscribed")
	time.Sleep(time.Second)
}
