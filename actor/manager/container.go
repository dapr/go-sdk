package manager

import (
	"context"
	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/codec"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"reflect"
)

type ActorContainer interface {
	Invoke(methodName string, param []byte) ([]reflect.Value, actorErr.ActorErr)
	GetActor() actor.Server
}

// DefaultActorContainer contains actor instance and methods type info generated from actor
type DefaultActorContainer struct {
	methodType map[string]*MethodType
	actor      actor.Server
	serializer codec.Codec
}

// NewDefaultActorContainer creates a new ActorContainer with provider impl actor and serializer
func NewDefaultActorContainer(actorID string, impl actor.Server, serializer codec.Codec) ActorContainer {
	impl.SetID(actorID)
	return &DefaultActorContainer{
		methodType: getAbsctractMethodMap(impl),
		actor:      impl,
		serializer: serializer,
	}
}

func (d *DefaultActorContainer) GetActor() actor.Server {
	return d.actor
}

// Invoke call actor method with given methodName and param
func (d *DefaultActorContainer) Invoke(methodName string, param []byte) ([]reflect.Value, actorErr.ActorErr) {
	methodType, ok := d.methodType[methodName]
	if !ok {
		return nil, actorErr.ErrActorMethodNoFound
	}
	argsValues := make([]reflect.Value, 0)
	argsValues = append(argsValues, reflect.ValueOf(d.actor))
	argsValues = append(argsValues, reflect.ValueOf(context.Background()))
	if len(methodType.argsType) > 0 {
		typ := methodType.argsType[0]
		paramValue := reflect.New(typ)
		paramInterface := paramValue.Interface()
		if err := d.serializer.Unmarshal(param, paramInterface); err != nil {
			return nil, actorErr.ErrActorSerializeFailed
		}
		argsValues = append(argsValues, reflect.ValueOf(paramInterface).Elem())
	}
	returnValue := methodType.method.Func.Call(argsValues)
	return returnValue, actorErr.Success
}
