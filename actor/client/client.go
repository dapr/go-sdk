package client

import (
	"context"
	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	"github.com/dapr/go-sdk/client"
)

type ActorClient interface {

	// RegisterActorTimer registers an actor timer.
	RegisterActorTimer(ctx context.Context, in *client.RegisterActorTimerRequest) error

	// UnregisterActorTimer unregisters an actor timer.
	UnregisterActorTimer(ctx context.Context, in *client.UnregisterActorTimerRequest) error

	// RegisterActorReminder registers an actor reminder.
	RegisterActorReminder(ctx context.Context, in *client.RegisterActorReminderRequest) error

	// UnregisterActorReminder unregisters an actor reminder.
	UnregisterActorReminder(ctx context.Context, in *client.UnregisterActorReminderRequest) error

	// InvokeActor calls a method on an actor.
	InvokeActor(ctx context.Context, in *client.InvokeActorRequest) (*client.InvokeActorResponse, error)

	// GetActorState get actor state
	GetActorState(ctx context.Context, in *client.GetActorStateRequest) (data *client.GetActorStateResponse, err error)

	// SaveStatetransactionally
	SaveStatetransactionally(ctx context.Context, actorType, actorID string, operations []*client.ActorStateOperation) error

	ImplActorClientStub(actorClientStub actor.Client, opt ...config.Option)
}
