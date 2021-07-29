package main

import (
	"context"
	"fmt"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/actor/api"
)

func main() {
	// just for this demo
	ctx := context.Background()

	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	myActor := new(api.ActorImpl)
	client.ImplActorInteface(myActor)
	usr, err := myActor.GetUser(ctx, &api.User{
		Name: "abc",
		Age:  123,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("get user = ", usr)

	rsp, err := myActor.Invoke(ctx, "laurence")
	if err != nil {
		panic(err)
	}
	fmt.Println("get invoke result = ", rsp)

	err = myActor.Post(ctx, "laurence")
	if err != nil {
		panic(err)
	}

	rsp, err = myActor.Get(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("get rsp = ", rsp)

	//err = client.RegisterActorTimer(ctx, &dapr.RegisterActorTimerRequest{
	//	ActorType: "testActorType",
	//	ActorID:   "testActorID",
	//	Name:      "testTimerName",
	//	DueTime:   "3s",
	//	Period:    "5s",
	//	Data:      []byte("hello"),
	//  Callback: "Invoke",
	//})
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("timer invoke set")

	//err = client.RegisterActorReminder(ctx, &dapr.RegisterActorReminderRequest{
	//	ActorType: "testActorType",
	//	ActorID:   "testActorID",
	//	Name:      "testTimerName",
	//	DueTime:   "3s",
	//	Period:    "5s",
	//	Data:      []byte("hello"),
	//})
}
