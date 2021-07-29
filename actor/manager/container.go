package manager

import "github.com/dapr/go-sdk/actor"

type ActorContainer struct {
	methodType map[string]*MethodType
	actor      actor.ActorImpl
}

func NewActorContainer(impl actor.ActorImpl) *ActorContainer {
	return &ActorContainer{
		methodType: getAbsctractMethodMap(impl),
		actor:      impl,
	}
}
