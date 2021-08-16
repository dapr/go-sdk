package mock

import (
	"context"
	"github.com/dapr/go-sdk/actor"
	dapr "github.com/dapr/go-sdk/client"
)

func MockActorImplFactory() actor.Server {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	return &MockActorImpl{
		daprClient: client,
	}
}

type MockActorImpl struct {
	actor.ServerImplBase
	daprClient dapr.Client
}

func (t *MockActorImpl) Type() string {
	return "testActorType"
}

func (t *MockActorImpl) Invoke(ctx context.Context, req string) (string, error) {
	return req, nil
}

func (t *MockActorImpl) ReminderCall(reminderName string, state []byte, dueTime string, period string) {
}

func MockNotReminderCalleeActorFactory() actor.Server {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	return &MockNotReminderCalleeActor{
		daprClient: client,
	}
}

type MockNotReminderCalleeActor struct {
	actor.ServerImplBase
	daprClient dapr.Client
}

func (t *MockNotReminderCalleeActor) Type() string {
	return "testActorNotReminderCalleeType"
}
