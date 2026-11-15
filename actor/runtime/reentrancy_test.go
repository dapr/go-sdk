/*
Copyright 2026 The Dapr Authors
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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/api"
	"github.com/dapr/go-sdk/actor/config"
	actorErr "github.com/dapr/go-sdk/actor/error"
	actorMock "github.com/dapr/go-sdk/actor/mock"
)

func TestConfigureReentrancyJSONClonesInputs(t *testing.T) {
	rt := NewActorRuntimeContext()
	depth := 7
	entities := []string{"alpha", "beta"}
	rt.ConfigureReentrancy(
		&config.ReentrancyConfig{Enabled: true, MaxStackDepth: &depth},
		[]config.EntityConfig{{Entities: entities}},
	)
	depth = 9
	entities[0] = "mutated"

	data, err := rt.GetJSONSerializedConfig()
	require.NoError(t, err)
	var cfg api.ActorRuntimeConfig
	require.NoError(t, json.Unmarshal(data, &cfg))
	require.NotNil(t, cfg.Reentrancy)
	assert.True(t, cfg.Reentrancy.Enabled)
	require.NotNil(t, cfg.Reentrancy.MaxStackDepth)
	assert.Equal(t, 7, *cfg.Reentrancy.MaxStackDepth)
	require.Len(t, cfg.EntityConfigs, 1)
	assert.Equal(t, []string{"alpha", "beta"}, cfg.EntityConfigs[0].Entities)
	assert.Nil(t, cfg.EntityConfigs[0].Reentrancy)

	rt.ConfigureReentrancy(nil, nil)
	data, err = rt.GetJSONSerializedConfig()
	require.NoError(t, err)
	cfg = api.ActorRuntimeConfig{}
	require.NoError(t, json.Unmarshal(data, &cfg))
	assert.Nil(t, cfg.Reentrancy)
	assert.Empty(t, cfg.EntityConfigs)
}

func TestConfigureReentrancyIgnoresMockManager(t *testing.T) {
	rt := NewActorRuntimeContext()
	ctrl := gomock.NewController(t)
	rt.actorManagers.Store("testActorType", actorMock.NewMockActorManagerContext(ctrl))
	rt.ConfigureReentrancy(&config.ReentrancyConfig{Enabled: true}, nil)
}

func TestConfigureReentrancyPrecedence(t *testing.T) {
	t.Setenv("DAPR_CLIENT_TIMEOUT_SECONDS", "-1")

	t.Run("global enables an unlisted type", func(t *testing.T) {
		assertReentrancy(t, "globalOn", true, &config.ReentrancyConfig{Enabled: true}, nil)
	})
	t.Run("nil entity reentrancy disables a listed type", func(t *testing.T) {
		assertReentrancy(t, "entityOff", false, &config.ReentrancyConfig{Enabled: true}, []config.EntityConfig{{
			Entities: []string{"entityOff"},
		}})
	})
	t.Run("entity enables when global is off", func(t *testing.T) {
		assertReentrancy(t, "entityOn", true, nil, []config.EntityConfig{{
			Entities:   []string{"entityOn"},
			Reentrancy: &config.ReentrancyConfig{Enabled: true},
		}})
	})
	t.Run("last entity entry wins", func(t *testing.T) {
		assertReentrancy(t, "lastWins", false, &config.ReentrancyConfig{Enabled: true}, []config.EntityConfig{
			{Entities: []string{"lastWins"}, Reentrancy: &config.ReentrancyConfig{Enabled: true}},
			{Entities: []string{"lastWins"}},
		})
	})
	t.Run("configured after registration", func(t *testing.T) {
		rt := NewActorRuntimeContext()
		release, wait, started := newOverlapGate()
		defer release()
		registerGateActor(rt, "lateConfig", started, wait)
		rt.ConfigureReentrancy(nil, []config.EntityConfig{{
			Entities:   []string{"lateConfig"},
			Reentrancy: &config.ReentrancyConfig{Enabled: true},
		}})
		assertOverlap(t, rt, "lateConfig", true, release, started)
	})
}

func TestActorTypesRunConcurrently(t *testing.T) {
	t.Setenv("DAPR_CLIENT_TIMEOUT_SECONDS", "-1")
	rt := NewActorRuntimeContext()
	started := make(chan string, 2)
	release, wait := closeRuntimeSignal()
	defer release()

	register := func(actorType string) {
		var once atomic.Bool
		rt.RegisterActorFactory(func() actor.ServerContext {
			return &GateActor{
				actorType: actorType,
				hold: func() {
					if once.CompareAndSwap(false, true) {
						started <- actorType
						<-wait
					}
				},
			}
		})
	}
	register("typeA")
	register("typeB")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = rt.InvokeActorMethod(context.Background(), "typeA", "id", "Hold", nil)
	}()
	go func() {
		defer wg.Done()
		_, _ = rt.InvokeActorMethod(context.Background(), "typeB", "id", "Hold", nil)
	}()

	got := map[string]bool{}
	deadline := time.After(5 * time.Second)
	for len(got) < 2 {
		select {
		case actorType := <-started:
			got[actorType] = true
		case <-deadline:
			t.Fatalf("actor types did not overlap: %v", got)
		}
	}
	release()
	wg.Wait()
}

type GateActor struct {
	actor.ServerImplBaseCtx
	actorType string
	hold      func()
}

func (a *GateActor) Type() string { return a.actorType }

func (a *GateActor) Hold(context.Context) error {
	if a.hold != nil {
		a.hold()
	}
	return nil
}

func (a *GateActor) Ping(context.Context) error { return nil }

func assertReentrancy(t *testing.T, actorType string, wantOverlap bool, global *config.ReentrancyConfig, entities []config.EntityConfig) {
	t.Helper()
	rt := NewActorRuntimeContext()
	release, wait, started := newOverlapGate()
	defer release()
	rt.ConfigureReentrancy(global, entities)
	registerGateActor(rt, actorType, started, wait)
	assertOverlap(t, rt, actorType, wantOverlap, release, started)
}

func registerGateActor(rt *ActorRunTimeContext, actorType string, started chan struct{}, wait <-chan struct{}) {
	var once atomic.Bool
	rt.RegisterActorFactory(func() actor.ServerContext {
		return &GateActor{
			actorType: actorType,
			hold: func() {
				if once.CompareAndSwap(false, true) {
					close(started)
					<-wait
				}
			},
		}
	})
}

func newOverlapGate() (func(), <-chan struct{}, chan struct{}) {
	release, wait := closeRuntimeSignal()
	return release, wait, make(chan struct{})
}

func assertOverlap(t *testing.T, rt *ActorRunTimeContext, actorType string, wantOverlap bool, release func(), started <-chan struct{}) {
	t.Helper()
	// The first call in the chain carries the id. Later calls may enter only
	// when that same id is still the active turn.
	chain := actor.WithReentrancyID(context.Background(), "chain")
	first := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := rt.InvokeActorMethod(chain, actorType, "id", "Hold", nil)
		first <- aerr
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("actor method did not start")
	}

	if wantOverlap {
		second := make(chan actorErr.ActorErr, 1)
		go func() {
			_, aerr := rt.InvokeActorMethod(chain, actorType, "id", "Ping", nil)
			second <- aerr
		}()
		require.Equal(t, actorErr.Success, recvRuntime(t, second))
		release()
		require.Equal(t, actorErr.Success, recvRuntime(t, first))
		return
	}

	// A queued call whose context ends returns without running Ping. Success
	// here would mean it joined the active turn.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	second := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := rt.InvokeActorMethod(actor.WithReentrancyID(ctx, "chain"), actorType, "id", "Ping", nil)
		second <- aerr
	}()
	require.Equal(t, actorErr.ErrActorTurnCanceled, recvRuntime(t, second))
	release()
	require.Equal(t, actorErr.Success, recvRuntime(t, first))
}

func closeRuntimeSignal() (func(), <-chan struct{}) {
	ch := make(chan struct{})
	var once sync.Once
	return func() { once.Do(func() { close(ch) }) }, ch
}

func recvRuntime[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting")
	}
	var zero T
	return zero
}
