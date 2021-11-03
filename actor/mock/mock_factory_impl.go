package mock

import (
	"context"

	"github.com/dapr/go-sdk/actor"
)

func ActorImplFactory() actor.Server {
	return &ActorImpl{}
}

type ActorImpl struct {
	actor.ServerImplBase
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
	return &NotReminderCalleeActor{}
}

type NotReminderCalleeActor struct {
	actor.ServerImplBase
}

func (t *NotReminderCalleeActor) Type() string {
	return "testActorNotReminderCalleeType"
}
