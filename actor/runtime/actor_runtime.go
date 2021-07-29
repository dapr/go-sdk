package runtime

import (
	"encoding/json"
	"github.com/dapr/go-sdk/actor"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/manager"
	"sync"
)

type ActorRunTime struct {
	config        ActorRuntimeConfig
	actorManagers sync.Map
}

var actorRuntimeInstance *ActorRunTime

func NewActorRuntime() *ActorRunTime {
	return &ActorRunTime{}
}

func GetActorRuntime() *ActorRunTime {
	if actorRuntimeInstance == nil {
		actorRuntimeInstance = &ActorRunTime{}
	}
	return actorRuntimeInstance
}

func (r *ActorRunTime) RegisterActorFactory(f actor.ActorImplFactory) {
	arctorType := f().Type()
	r.config.RegisteredActorTypes = append(r.config.RegisteredActorTypes, arctorType)
	mng, ok := r.actorManagers.Load(arctorType)
	if !ok {
		newMng := manager.NewActorManager()
		newMng.RegisterActorImplFactory(f)
		r.actorManagers.Store(arctorType, newMng)
		return
	}
	mng.(*manager.ActorManager).RegisterActorImplFactory(f)
}

func (r *ActorRunTime) GetSerializedConfig() []byte {
	data, _ := json.Marshal(&r.config)
	return data
}

func (r *ActorRunTime) InvokeActorMethod(actorTypeName, actorID, actorMethod string, payload []byte) ([]byte, actorErr.ActorError) {
	mng, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return nil, actorErr.ErrorActorTypeNotFound
	}
	return mng.(*manager.ActorManager).InvokeMethod(actorID, actorMethod, payload)
}

func (r *ActorRunTime) Deactive(actorTypeName, actorID string) actorErr.ActorError {
	targetManager, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return actorErr.ErrorActorTypeNotFound
	}
	return targetManager.(*manager.ActorManager).DetectiveActor(actorID)
}

func (r *ActorRunTime) InvokeReminder(actorTypeName, actorID, reminderName string, params []byte) actorErr.ActorError {
	targetManager, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return actorErr.ErrorActorTypeNotFound
	}
	mng := targetManager.(*manager.ActorManager)
	mng.ActiveManager(actorID)
	return mng.InvokeReminder(actorID, reminderName, params)
}

func (r *ActorRunTime) InvokeTimer(actorTypeName, actorID, timerName string, params []byte) actorErr.ActorError {
	targetManager, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return actorErr.ErrorActorTypeNotFound
	}
	mng := targetManager.(*manager.ActorManager)
	mng.ActiveManager(actorID)
	return mng.InvokeTimer(actorID, timerName, params)
}
