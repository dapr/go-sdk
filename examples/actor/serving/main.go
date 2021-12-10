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
	"log"
	"net/http"

	"github.com/dapr/go-sdk/actor"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/actor/api"

	daprd "github.com/dapr/go-sdk/service/http"
)

func testActorFactory() actor.Server {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	return &TestActor{
		daprClient: client,
	}
}

type TestActor struct {
	actor.ServerImplBase
	daprClient dapr.Client
}

func (t *TestActor) Type() string {
	return "testActorType"
}

// user defined functions
func (t *TestActor) StopTimer(ctx context.Context, req *api.TimerRequest) error {
	return t.daprClient.UnregisterActorTimer(ctx, &dapr.UnregisterActorTimerRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.TimerName,
	})
}

func (t *TestActor) StartTimer(ctx context.Context, req *api.TimerRequest) error {
	return t.daprClient.RegisterActorTimer(ctx, &dapr.RegisterActorTimerRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.TimerName,
		DueTime:   req.Duration,
		Period:    req.Period,
		Data:      []byte(req.Data),
		CallBack:  req.CallBack,
	})
}

func (t *TestActor) StartReminder(ctx context.Context, req *api.ReminderRequest) error {
	return t.daprClient.RegisterActorReminder(ctx, &dapr.RegisterActorReminderRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.ReminderName,
		DueTime:   req.Duration,
		Period:    req.Period,
		Data:      []byte(req.Data),
	})
}

func (t *TestActor) StopReminder(ctx context.Context, req *api.ReminderRequest) error {
	return t.daprClient.UnregisterActorReminder(ctx, &dapr.UnregisterActorReminderRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.ReminderName,
	})
}

func (t *TestActor) Invoke(ctx context.Context, req string) (string, error) {
	fmt.Println("get req = ", req)
	return req, nil
}

func (t *TestActor) GetUser(ctx context.Context, user *api.User) (*api.User, error) {
	fmt.Println("call get user req = ", user)
	return user, nil
}

func (t *TestActor) Get(context.Context) (string, error) {
	return "get result", nil
}

func (t *TestActor) Post(ctx context.Context, req string) error {
	fmt.Println("get post request = ", req)
	return nil
}

func (t *TestActor) IncrementAndGet(ctx context.Context, stateKey string) (*api.User, error) {
	stateData := api.User{}
	if exist, err := t.GetStateManager().Contains(stateKey); err != nil {
		fmt.Println("state manager call contains with key " + stateKey + "err = " + err.Error())
		return &stateData, err
	} else if exist {
		if err := t.GetStateManager().Get(stateKey, &stateData); err != nil {
			fmt.Println("state manager call get with key " + stateKey + "err = " + err.Error())
			return &stateData, err
		}
	}
	stateData.Age++
	if err := t.GetStateManager().Set(stateKey, stateData); err != nil {
		fmt.Printf("state manager set get with key %s and state data = %+v, error = %s", stateKey, stateData, err.Error())
		return &stateData, err
	}
	return &stateData, nil
}

func (t *TestActor) ReminderCall(reminderName string, state []byte, dueTime string, period string) {
	fmt.Println("receive reminder = ", reminderName, " state = ", string(state), "duetime = ", dueTime, "period = ", period)
}

func main() {
	s := daprd.NewService(":8080")
	s.RegisterActorImplFactory(testActorFactory)
	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}
