package main

import (
	"context"
	"fmt"
	"github.com/dapr/go-sdk/actor"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/actor/api"
	"log"
	"net/http"

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
