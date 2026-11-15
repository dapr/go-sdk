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

package runtime

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/api"
	"github.com/dapr/go-sdk/actor/config"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/manager"
)

// Deprecated: use ActorRunTimeContext instead.
type ActorRunTime struct {
	ctx *ActorRunTimeContext
}

type ActorRunTimeContext struct {
	mu            sync.Mutex
	config        api.ActorRuntimeConfig
	actorManagers sync.Map
}

var (
	actorRuntimeInstance    *ActorRunTime
	actorRuntimeInstanceCtx *ActorRunTimeContext
)

// NewActorRuntime creates an empty ActorRuntime.
//
// Deprecated: use NewActorRuntimeContext instead.
func NewActorRuntime() *ActorRunTime {
	return &ActorRunTime{ctx: NewActorRuntimeContext()}
}

// NewActorRuntimeContext creates an empty ActorRuntimeContext.
func NewActorRuntimeContext() *ActorRunTimeContext {
	return &ActorRunTimeContext{}
}

// GetActorRuntimeInstance gets or create runtime instance.
//
// Deprecated: use GetActorRuntimeInstanceContext instead.
func GetActorRuntimeInstance() *ActorRunTime {
	if actorRuntimeInstance == nil {
		actorRuntimeInstance = NewActorRuntime()
	}
	return actorRuntimeInstance
}

// GetActorRuntimeInstanceContext gets or create runtime instance.
func GetActorRuntimeInstanceContext() *ActorRunTimeContext {
	if actorRuntimeInstanceCtx == nil {
		actorRuntimeInstanceCtx = NewActorRuntimeContext()
	}
	return actorRuntimeInstanceCtx
}

// RegisterActorFactory registers the given actor factory from user, and create new actor manager if not exists.
func (r *ActorRunTimeContext) RegisterActorFactory(f actor.FactoryContext, opt ...config.Option) {
	conf := config.GetConfigFromOptions(opt...)
	actType := f().Type()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config.RegisteredActorTypes = append(r.config.RegisteredActorTypes, actType)
	enabled := config.ReentrancyEnabled(actType, r.config.Reentrancy, r.config.EntityConfigs)
	mng, ok := r.actorManagers.Load(actType)
	if !ok {
		newMng, err := manager.NewDefaultActorManagerContext(conf.SerializerType)
		if err != actorErr.Success {
			return
		}
		newMng.RegisterActorImplFactory(f)
		setManagerReentrancy(newMng, enabled)
		r.actorManagers.Store(actType, newMng)
		return
	}
	mng.(manager.ActorManagerContext).RegisterActorImplFactory(f)
	setManagerReentrancy(mng.(manager.ActorManagerContext), enabled)
}

// ConfigureReentrancy applies the reentrancy settings daprd was given.
// A per-entity entry replaces the global setting for the actor types it names,
// including disabling reentrancy when that entry has no reentrancy block.
// The same values are returned from GET /dapr/config.
func (r *ActorRunTimeContext) ConfigureReentrancy(global *config.ReentrancyConfig, entities []config.EntityConfig) {
	global = cloneReentrancy(global)
	entities = cloneEntityConfigs(entities)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config.Reentrancy = global
	r.config.EntityConfigs = entities

	// Hold mu across the update so a factory registered in between cannot keep
	// a flag computed from the previous config.
	r.actorManagers.Range(func(key, value any) bool {
		mng, ok := value.(manager.ActorManagerContext)
		if ok {
			setManagerReentrancy(mng, config.ReentrancyEnabled(key.(string), global, entities))
		}
		return true
	})
}

func setManagerReentrancy(mng manager.ActorManagerContext, enabled bool) {
	if setter, ok := mng.(*manager.DefaultActorManagerContext); ok {
		setter.SetReentrancyEnabled(enabled)
	}
}

func cloneReentrancy(in *config.ReentrancyConfig) *config.ReentrancyConfig {
	if in == nil {
		return nil
	}
	out := *in
	if in.MaxStackDepth != nil {
		depth := *in.MaxStackDepth
		out.MaxStackDepth = &depth
	}
	return &out
}

func cloneEntityConfigs(in []config.EntityConfig) []config.EntityConfig {
	if len(in) == 0 {
		return nil
	}
	out := make([]config.EntityConfig, len(in))
	for i, entity := range in {
		out[i] = config.EntityConfig{
			Entities:   append([]string(nil), entity.Entities...),
			Reentrancy: cloneReentrancy(entity.Reentrancy),
		}
	}
	return out
}

func (r *ActorRunTimeContext) GetJSONSerializedConfig() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return json.Marshal(&r.config)
}

// RegisteredActorTypes returns the names of the actor types registered on
// this runtime through RegisterActorFactory.
func (r *ActorRunTimeContext) RegisteredActorTypes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	types := make([]string, len(r.config.RegisteredActorTypes))
	copy(types, r.config.RegisteredActorTypes)
	return types
}

func (r *ActorRunTimeContext) InvokeActorMethod(ctx context.Context, actorTypeName, actorID, actorMethod string, payload []byte) ([]byte, actorErr.ActorErr) {
	mng, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return nil, actorErr.ErrActorTypeNotFound
	}
	return mng.(manager.ActorManagerContext).InvokeMethod(ctx, actorID, actorMethod, payload)
}

func (r *ActorRunTimeContext) Deactivate(ctx context.Context, actorTypeName, actorID string) actorErr.ActorErr {
	targetManager, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return actorErr.ErrActorTypeNotFound
	}
	return targetManager.(manager.ActorManagerContext).DeactivateActor(ctx, actorID)
}

func (r *ActorRunTimeContext) InvokeReminder(ctx context.Context, actorTypeName, actorID, reminderName string, params []byte) actorErr.ActorErr {
	targetManager, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return actorErr.ErrActorTypeNotFound
	}
	mng := targetManager.(manager.ActorManagerContext)
	return mng.InvokeReminder(ctx, actorID, reminderName, params)
}

func (r *ActorRunTimeContext) InvokeTimer(ctx context.Context, actorTypeName, actorID, timerName string, params []byte) actorErr.ActorErr {
	targetManager, ok := r.actorManagers.Load(actorTypeName)
	if !ok {
		return actorErr.ErrActorTypeNotFound
	}
	mng := targetManager.(manager.ActorManagerContext)
	return mng.InvokeTimer(ctx, actorID, timerName, params)
}

// Deprecated: use ActorRunTimeContext instead.
func (r *ActorRunTime) RegisterActorFactory(f actor.Factory, opt ...config.Option) {
	r.ctx.RegisterActorFactory(func() actor.ServerContext { return f().WithContext() }, opt...)
}

// Deprecated: use ActorRunTimeContext instead.
func (r *ActorRunTime) GetJSONSerializedConfig() ([]byte, error) {
	return r.ctx.GetJSONSerializedConfig()
}

// Deprecated: use ActorRunTimeContext instead.
func (r *ActorRunTime) InvokeActorMethod(actorTypeName, actorID, actorMethod string, payload []byte) ([]byte, actorErr.ActorErr) {
	return r.ctx.InvokeActorMethod(context.Background(), actorTypeName, actorID, actorMethod, payload)
}

// Deprecated: use ActorRunTimeContext instead.
func (r *ActorRunTime) Deactivate(actorTypeName, actorID string) actorErr.ActorErr {
	return r.ctx.Deactivate(context.Background(), actorTypeName, actorID)
}

// Deprecated: use ActorRunTimeContext instead.
func (r *ActorRunTime) InvokeReminder(actorTypeName, actorID, reminderName string, params []byte) actorErr.ActorErr {
	return r.ctx.InvokeReminder(context.Background(), actorTypeName, actorID, reminderName, params)
}

// Deprecated: use ActorRunTimeContext instead.
func (r *ActorRunTime) InvokeTimer(actorTypeName, actorID, timerName string, params []byte) actorErr.ActorErr {
	return r.ctx.InvokeTimer(context.Background(), actorTypeName, actorID, timerName, params)
}
