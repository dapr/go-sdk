package mock

import (
	"context"

	"github.com/dapr/go-sdk/actor"
	dapr "github.com/dapr/go-sdk/client"
)

func ActorImplFactory() actor.Server {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	return &ActorImpl{
		daprClient: client,
	}
}

type ActorImpl struct {
	actor.ServerImplBase
	daprClient dapr.Client
}

func (t *ActorImpl) Type() string {
	return "testActorType"
}

func (t *ActorImpl) Invoke(ctx context.Context, req string) (string, error) {
	return req, nil
}

func (t *ActorImpl) ReminderCall(reminderName string, state []byte, dueTime string, period string) {
}

func NotReminderCalleeActorFactory() actor.Server {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	return &NotReminderCalleeActor{
		daprClient: client,
	}
}

type NotReminderCalleeActor struct {
	actor.ServerImplBase
	daprClient dapr.Client
}

func (t *NotReminderCalleeActor) Type() string {
	return "testActorNotReminderCalleeType"
}
