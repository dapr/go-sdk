package main

import (
	"context"
	"fmt"
	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/examples/actor/api"
	"log"
	"net/http"

	daprd "github.com/dapr/go-sdk/service/http"
)

func testActorFactory() actor.ActorImpl {
	return &TestActor{}
}

type TestActor struct {
}

func (t *TestActor) OnDeactive() {
	panic("implement me")
}

func (t *TestActor) OnActive() {
	panic("implement me")
}

func (t *TestActor) Type() string {
	return "testActorType"
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

func (t *TestActor) ReceiveReminder(reminderName string, state interface{}, dueTime string, period string) []byte {
	fmt.Println("receive reminder = ", reminderName, " state = ", state, "duetime = ", dueTime, "period = ", period)
	return nil
}

func main() {
	s := daprd.NewService(":18080")

	s.RegisterActorImplFactory(testActorFactory)

	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}
