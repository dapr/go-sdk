package main

import (
	"context"
	"fmt"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/actor/api"
)

func main() {
	ctx := context.Background()

	// create the client
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// implement actor client stub
	myActor := new(api.ClientStub)
	client.ImplActorClientStub(myActor)

	// Invoke user defined method GetUser with user defined param api.User and response
	// using default serializer type json
	user, err := myActor.GetUser(ctx, &api.User{
		Name: "abc",
		Age:  123,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("get user result = ", user)

	// Invoke user defined method Invoke
	rsp, err := myActor.Invoke(ctx, "laurence")
	if err != nil {
		panic(err)
	}
	fmt.Println("get invoke result = ", rsp)

	// Invoke user defined method Post with empty response
	err = myActor.Post(ctx, "laurence")
	if err != nil {
		panic(err)
	}
	fmt.Println("get post result = ", rsp)

	// Invoke user defined method Get with empty request param
	rsp, err = myActor.Get(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("get result = ", rsp)

	// Invoke user defined method StarTimer, and server side actor start actor timer with given params.
	err = myActor.StartTimer(ctx, &api.TimerRequest{
		TimerName: "testTimerName",
		CallBack:  "Invoke",
		Period:    "5s",
		Duration:  "5s",
		Data:      `"hello"`,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("start timer")
	<-time.After(time.Second * 10) // timer call for twice

	// Invoke user defined method StopTimer, and server side actor stop actor timer with given params.
	err = myActor.StopTimer(ctx, &api.TimerRequest{
		TimerName: "testTimerName",
		CallBack:  "Invoke",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("stop timer")

	// Invoke user defined method StartReminder, and server side actor start actor reminder with given params.
	err = myActor.StartReminder(ctx, &api.ReminderRequest{
		ReminderName: "testReminderName",
		Period:       "5s",
		Duration:     "5s",
		Data:         `"hello"`,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("start reminder")
	<-time.After(time.Second * 10) // timer call for twice

	// Invoke user defined method StopReminder, and server side actor stop actor reminder with given params.
	err = myActor.StopReminder(ctx, &api.ReminderRequest{
		ReminderName: "testReminderName",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("stop reminder")

	for i := 0; i < 2; i++ {
		// Invoke user defined method IncrementAndGet, and server side actor increase the state named testStateKey and return.
		usr, err := myActor.IncrementAndGet(ctx, "testStateKey")
		if err != nil {
			panic(err)
		}
		fmt.Printf("get user = %+v\n", *usr)
		time.Sleep(time.Second)
	}
}
