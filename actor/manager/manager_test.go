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

package manager

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/api"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/mock"
)

func TestMain(m *testing.M) {
	// no sidecar in unit tests, fail the dapr client dial fast instead of blocking 5s per new actor
	os.Setenv("DAPR_CLIENT_TIMEOUT_SECONDS", "-1")
	os.Exit(m.Run())
}

func TestNewDefaultActorManager(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)

	mng, err = NewDefaultActorManager("badSerializerType")
	require.NotNil(t, mng)
	require.Nil(t, mng.(*DefaultActorManager).ctx)
	assert.Equal(t, actorErr.ErrActorSerializeNoFound, err)
}

func TestRegisterActorImplFactory(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	require.NotNil(t, mng)
	require.Nil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)
	assert.Equal(t, actorErr.Success, err)
	mng.RegisterActorImplFactory(mock.ActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)
}

func TestInvokeMethod(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)

	data, err := mng.InvokeMethod("testActorID", "testMethodName", []byte(`"hello"`))
	assert.Nil(t, data)
	assert.Equal(t, actorErr.ErrActorFactoryNotSet, err)

	mng.RegisterActorImplFactory(mock.ActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)
	data, err = mng.InvokeMethod("testActorID", "mockMethod", []byte(`"hello"`))
	assert.Nil(t, data)
	assert.Equal(t, actorErr.ErrActorMethodNoFound, err)

	data, err = mng.InvokeMethod("testActorID", "Invoke", []byte(`"hello"`))
	assert.Equal(t, data, []byte(`"hello"`))
	assert.Equal(t, actorErr.Success, err)
}

func TestDeactivateActor(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)

	err = mng.DeactivateActor("testActorID")
	assert.Equal(t, actorErr.ErrActorIDNotFound, err)

	mng.RegisterActorImplFactory(mock.ActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)
	mng.InvokeMethod("testActorID", "Invoke", []byte(`"hello"`))

	err = mng.DeactivateActor("testActorID")
	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeReminder(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)

	err = mng.InvokeReminder("testActorID", "testReminderName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrActorFactoryNotSet, err)

	mng.RegisterActorImplFactory(mock.ActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)
	err = mng.InvokeReminder("testActorID", "testReminderName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrRemindersParamsInvalid, err)

	reminderParam, _ := json.Marshal(&api.ActorReminderParams{
		Data:    []byte("hello"),
		DueTime: "5s",
		Period:  "6s",
	})
	err = mng.InvokeReminder("testActorID", "testReminderName", reminderParam)
	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeTimer(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)

	err = mng.InvokeTimer("testActorID", "testTimerName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrActorFactoryNotSet, err)

	mng.RegisterActorImplFactory(mock.ActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).ctx.(*DefaultActorManagerContext).factory)
	err = mng.InvokeTimer("testActorID", "testTimerName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrTimerParamsInvalid, err)

	timerParam, _ := json.Marshal(&api.ActorTimerParam{
		Data:     []byte("hello"),
		DueTime:  "5s",
		Period:   "6s",
		CallBack: "Invoke",
	})
	err = mng.InvokeTimer("testActorID", "testTimerName", timerParam)
	assert.Equal(t, actorErr.ErrActorMethodSerializeFailed, err)

	timerParam, _ = json.Marshal(&api.ActorTimerParam{
		Data:     []byte("hello"),
		DueTime:  "5s",
		Period:   "6s",
		CallBack: "NoSuchMethod",
	})
	err = mng.InvokeTimer("testActorID", "testTimerName", timerParam)
	assert.Equal(t, actorErr.ErrActorMethodNoFound, err)

	timerParam, _ = json.Marshal(&api.ActorTimerParam{
		Data:     []byte(`"hello"`),
		DueTime:  "5s",
		Period:   "6s",
		CallBack: "Invoke",
	})
	err = mng.InvokeTimer("testActorID", "testTimerName", timerParam)
	assert.Equal(t, actorErr.Success, err)
}

type ActivationTestActor struct {
	actor.ServerImplBaseCtx
	count atomic.Int64
	hold  func()
}

func (a *ActivationTestActor) Type() string {
	return "activationTestActor"
}

func (a *ActivationTestActor) Count(context.Context) (int64, error) {
	return a.count.Add(1), nil
}

func (a *ActivationTestActor) Hold(context.Context) (int64, error) {
	if a.hold != nil {
		a.hold()
	}
	return a.count.Add(1), nil
}

func (a *ActivationTestActor) ReminderCall(string, []byte, string, string) {
	a.count.Add(1)
}

type activationTestManager struct {
	*DefaultActorManagerContext
	lock    sync.Mutex
	created []*ActivationTestActor
}

func newActivationTestManager(t *testing.T) *activationTestManager {
	t.Helper()
	mng, aerr := NewDefaultActorManagerContext("json")
	require.Equal(t, actorErr.Success, aerr)
	m := &activationTestManager{DefaultActorManagerContext: mng.(*DefaultActorManagerContext)}
	m.RegisterActorImplFactory(func() actor.ServerContext {
		a := &ActivationTestActor{}
		m.lock.Lock()
		m.created = append(m.created, a)
		m.lock.Unlock()
		return a
	})
	return m
}

func (m *activationTestManager) createdCount() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.created)
}

func (m *activationTestManager) live(t *testing.T, actorID string) (*ActivationTestActor, string) {
	t.Helper()
	val, ok := m.activeActors.Load(actorID)
	require.True(t, ok)
	active := val.(*activeActor)
	return active.container.GetActor().(*ActivationTestActor), active.activationID
}

func (m *activationTestManager) count(t *testing.T, ctx context.Context, actorID string) int64 {
	t.Helper()
	data, aerr := m.InvokeMethod(ctx, actorID, "Count", nil)
	require.Equal(t, actorErr.Success, aerr)
	var n int64
	require.NoError(t, json.Unmarshal(data, &n))
	return n
}

func (m *activationTestManager) reminder(t *testing.T, ctx context.Context, actorID string) {
	t.Helper()
	params, err := json.Marshal(&api.ActorReminderParams{DueTime: "1s", Period: "1s"})
	require.NoError(t, err)
	require.Equal(t, actorErr.Success, m.InvokeReminder(ctx, actorID, "r", params))
}

func (m *activationTestManager) timer(t *testing.T, ctx context.Context, actorID string) {
	t.Helper()
	params, err := json.Marshal(&api.ActorTimerParam{CallBack: "Count", DueTime: "1s", Period: "1s"})
	require.NoError(t, err)
	require.Equal(t, actorErr.Success, m.InvokeTimer(ctx, actorID, "t", params))
}

func TestActivationIDSameIDReusesActor(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	ctx := actor.WithActivationID(t.Context(), "a")

	assert.Equal(t, int64(1), m.count(t, ctx, "id"))
	first, _ := m.live(t, "id")
	assert.Equal(t, int64(2), m.count(t, ctx, "id"))

	second, activationID := m.live(t, "id")
	assert.Same(t, first, second)
	assert.Equal(t, "a", activationID)
	assert.Equal(t, 1, m.createdCount())
}

func TestActivationIDNewIDResetsActor(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
	assert.Equal(t, int64(2), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
	old, _ := m.live(t, "id")
	oldStateManager := old.GetStateManager()

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "b"), "id"))
	fresh, activationID := m.live(t, "id")
	assert.NotSame(t, old, fresh)
	assert.NotSame(t, oldStateManager, fresh.GetStateManager())
	assert.Equal(t, "b", activationID)
	assert.Equal(t, int64(2), old.count.Load())
	assert.Equal(t, 2, m.createdCount())

	assert.Equal(t, int64(2), m.count(t, actor.WithActivationID(t.Context(), "b"), "id"))
	assert.Equal(t, 2, m.createdCount())
}

func TestActivationIDRetiredIDRejected(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "b"), "id"))
	live, _ := m.live(t, "id")

	_, aerr := m.InvokeMethod(actor.WithActivationID(t.Context(), "a"), "id", "Count", nil)
	assert.Equal(t, actorErr.ErrActorIDNotFound, aerr)
	params, err := json.Marshal(&api.ActorReminderParams{DueTime: "1s", Period: "1s"})
	require.NoError(t, err)
	assert.Equal(t, actorErr.ErrActorIDNotFound, m.InvokeReminder(actor.WithActivationID(t.Context(), "a"), "id", "r", params))

	after, activationID := m.live(t, "id")
	assert.Same(t, live, after)
	assert.Equal(t, "b", activationID)
	assert.Equal(t, int64(1), live.count.Load())
	assert.Equal(t, 2, m.createdCount())

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "c"), "id"))
	assert.Equal(t, 3, m.createdCount())
	_, aerr = m.InvokeMethod(actor.WithActivationID(t.Context(), "b"), "id", "Count", nil)
	assert.Equal(t, actorErr.ErrActorIDNotFound, aerr)
	assert.Equal(t, int64(2), m.count(t, actor.WithActivationID(t.Context(), "c"), "id"))
	assert.Equal(t, 3, m.createdCount())
}

func TestActivationIDReplacedCallSkipsSave(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
	old, _ := m.live(t, "id")
	held := make(chan struct{})
	release := make(chan struct{})
	old.hold = func() {
		close(held)
		<-release
	}

	done := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(actor.WithActivationID(t.Context(), "a"), "id", "Hold", nil)
		done <- aerr
	}()
	<-held

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "b"), "id"))
	close(release)
	assert.Equal(t, actorErr.ErrSaveStateFailed, <-done)

	fresh, activationID := m.live(t, "id")
	assert.NotSame(t, old, fresh)
	assert.Equal(t, "b", activationID)
	assert.Equal(t, int64(2), m.count(t, actor.WithActivationID(t.Context(), "b"), "id"))
}

func TestActivationIDEmptyNeverResets(t *testing.T) {
	t.Parallel()

	t.Run("stored id kept on empty id", func(t *testing.T) {
		t.Parallel()
		m := newActivationTestManager(t)
		assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
		assert.Equal(t, int64(2), m.count(t, t.Context(), "id"))
		assert.Equal(t, int64(3), m.count(t, actor.WithActivationID(t.Context(), ""), "id"))
		_, activationID := m.live(t, "id")
		assert.Equal(t, "a", activationID)
		assert.Equal(t, 1, m.createdCount())
		assert.Equal(t, int64(4), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
		assert.Equal(t, 1, m.createdCount())
	})

	t.Run("no id at all", func(t *testing.T) {
		t.Parallel()
		m := newActivationTestManager(t)
		for i := int64(1); i <= 3; i++ {
			assert.Equal(t, i, m.count(t, t.Context(), "id"))
		}
		_, activationID := m.live(t, "id")
		assert.Empty(t, activationID)
		assert.Equal(t, 1, m.createdCount())
	})
}

func TestActivationIDStoredEmptyThenNonEmptyResets(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)

	assert.Equal(t, int64(1), m.count(t, t.Context(), "id"))
	assert.Equal(t, int64(2), m.count(t, t.Context(), "id"))
	old, _ := m.live(t, "id")

	assert.Equal(t, int64(1), m.count(t, actor.WithActivationID(t.Context(), "a"), "id"))
	fresh, activationID := m.live(t, "id")
	assert.NotSame(t, old, fresh)
	assert.Equal(t, "a", activationID)
	assert.Equal(t, 2, m.createdCount())
}

func TestActivationIDReminder(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)

	m.reminder(t, actor.WithActivationID(t.Context(), "a"), "id")
	m.reminder(t, actor.WithActivationID(t.Context(), "a"), "id")
	old, _ := m.live(t, "id")
	assert.Equal(t, int64(2), old.count.Load())
	assert.Equal(t, 1, m.createdCount())

	m.reminder(t, actor.WithActivationID(t.Context(), "b"), "id")
	fresh, activationID := m.live(t, "id")
	assert.NotSame(t, old, fresh)
	assert.Equal(t, int64(1), fresh.count.Load())
	assert.Equal(t, "b", activationID)

	m.reminder(t, t.Context(), "id")
	same, _ := m.live(t, "id")
	assert.Same(t, fresh, same)
	assert.Equal(t, 2, m.createdCount())
}

func TestActivationIDTimer(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)

	m.timer(t, actor.WithActivationID(t.Context(), "a"), "id")
	m.timer(t, actor.WithActivationID(t.Context(), "a"), "id")
	old, _ := m.live(t, "id")
	assert.Equal(t, int64(2), old.count.Load())
	assert.Equal(t, 1, m.createdCount())

	m.timer(t, actor.WithActivationID(t.Context(), "b"), "id")
	fresh, activationID := m.live(t, "id")
	assert.NotSame(t, old, fresh)
	assert.Equal(t, int64(1), fresh.count.Load())
	assert.Equal(t, "b", activationID)

	m.timer(t, t.Context(), "id")
	assert.Equal(t, int64(2), fresh.count.Load())
	assert.Equal(t, 2, m.createdCount())
}

func TestActivationIDDeactivateUnchanged(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	ctx := actor.WithActivationID(t.Context(), "a")

	assert.Equal(t, int64(1), m.count(t, ctx, "id"))
	assert.Equal(t, actorErr.Success, m.DeactivateActor(ctx, "id"))
	assert.Equal(t, actorErr.ErrActorIDNotFound, m.DeactivateActor(ctx, "id"))

	assert.Equal(t, int64(1), m.count(t, ctx, "id"))
	assert.Equal(t, 2, m.createdCount())
}

func TestActivationIDConcurrent(t *testing.T) {
	t.Parallel()

	t.Run("same id reuses one actor", func(t *testing.T) {
		t.Parallel()
		m := newActivationTestManager(t)
		ctx := actor.WithActivationID(t.Context(), "a")
		m.count(t, ctx, "id")

		const calls = 50
		var wg sync.WaitGroup
		for range calls {
			wg.Go(func() {
				_, aerr := m.InvokeMethod(ctx, "id", "Count", nil)
				assert.Equal(t, actorErr.Success, aerr)
			})
		}
		wg.Wait()

		live, _ := m.live(t, "id")
		assert.Equal(t, int64(calls+1), live.count.Load())
		assert.Equal(t, 1, m.createdCount())
	})

	t.Run("resets on many actors", func(t *testing.T) {
		t.Parallel()
		m := newActivationTestManager(t)

		const actors = 50
		var wg sync.WaitGroup
		for i := range actors {
			wg.Go(func() {
				actorID := strconv.Itoa(i)
				for _, id := range []string{"a", "a", "b"} {
					_, aerr := m.InvokeMethod(actor.WithActivationID(t.Context(), id), actorID, "Count", nil)
					assert.Equal(t, actorErr.Success, aerr)
				}
			})
		}
		wg.Wait()

		for i := range actors {
			live, activationID := m.live(t, strconv.Itoa(i))
			assert.Equal(t, int64(1), live.count.Load())
			assert.Equal(t, "b", activationID)
		}
		assert.Equal(t, 2*actors, m.createdCount())
	})
}
