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

package manager

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/api"
	actorErr "github.com/dapr/go-sdk/actor/error"
)

const turnTestTimeout = 5 * time.Second

func TestTurnLockDifferentActorsDoNotBlock(t *testing.T) {
	t.Parallel()
	l := newTurnLock()
	first, err := l.acquire(context.Background(), "a", "")
	require.NoError(t, err)
	secondDone := make(chan *turnLease, 1)
	go func() {
		lease, acquireErr := l.acquire(context.Background(), "b", "")
		if acquireErr != nil {
			secondDone <- nil
			return
		}
		secondDone <- lease
	}()
	second := recvWithin(t, secondDone)
	require.NotNil(t, second)
	first.release()
	second.release()
	assert.Zero(t, l.len())
}

func TestTurnLockAlreadyCanceledDoesNotAcquire(t *testing.T) {
	t.Parallel()
	l := newTurnLock()
	owner, err := l.acquire(context.Background(), "a", "r")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lease, err := l.acquire(ctx, "a", "r")
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, lease)

	active, waiters, ok := l.state("a")
	require.True(t, ok)
	assert.Equal(t, 1, active)
	assert.Zero(t, waiters)

	owner.release()
	assert.Zero(t, l.len())
}

func TestTurnLockCancelQueuedWaiterHandsTurnToNext(t *testing.T) {
	t.Parallel()
	l := newTurnLock()
	owner, err := l.acquire(context.Background(), "a", "")
	require.NoError(t, err)

	bctx, bcancel := context.WithCancel(context.Background())
	bDone := make(chan error, 1)
	cDone := make(chan *turnLease, 1)
	go func() {
		lease, acquireErr := l.acquire(bctx, "a", "")
		bDone <- acquireErr
		if acquireErr == nil {
			lease.release()
		}
	}()
	go func() {
		lease, acquireErr := l.acquire(context.Background(), "a", "")
		if acquireErr != nil {
			cDone <- nil
			return
		}
		cDone <- lease
	}()
	require.Eventually(t, func() bool {
		_, waiters, ok := l.state("a")
		return ok && waiters == 2
	}, turnTestTimeout, time.Millisecond)

	bcancel()
	require.ErrorIs(t, recvWithin(t, bDone), context.Canceled)
	active, waiters, ok := l.state("a")
	require.True(t, ok)
	assert.Equal(t, 1, active)
	assert.Equal(t, 1, waiters)

	owner.release()
	next := recvWithin(t, cDone)
	require.NotNil(t, next)
	active, waiters, ok = l.state("a")
	require.True(t, ok)
	assert.Equal(t, 1, active)
	assert.Zero(t, waiters)
	next.release()
	assert.Zero(t, l.len())
}

func TestTurnLockReentrancyWaitsUntilAllowed(t *testing.T) {
	t.Parallel()
	l := newTurnLock()
	owner, err := l.acquire(context.Background(), "a", "chain")
	require.NoError(t, err)

	entered := make(chan *turnLease, 1)
	go func() {
		lease, acquireErr := l.acquire(context.Background(), "a", "chain")
		if acquireErr != nil {
			entered <- nil
			return
		}
		entered <- lease
	}()
	require.Eventually(t, func() bool {
		_, waiters, ok := l.state("a")
		return ok && waiters == 1
	}, turnTestTimeout, time.Millisecond)
	select {
	case <-entered:
		t.Fatal("matching reentrancy id entered before AllowReentrancy")
	default:
	}

	owner.allowReentrancy()
	nested := recvWithin(t, entered)
	require.NotNil(t, nested)
	active, waiters, ok := l.state("a")
	require.True(t, ok)
	assert.Equal(t, 2, active)
	assert.Zero(t, waiters)

	// A different id still waits, and stays waiting if the nested call finishes first.
	foreign := make(chan *turnLease, 1)
	go func() {
		lease, acquireErr := l.acquire(context.Background(), "a", "other")
		if acquireErr != nil {
			foreign <- nil
			return
		}
		foreign <- lease
	}()
	require.Eventually(t, func() bool {
		_, queued, present := l.state("a")
		return present && queued == 1
	}, turnTestTimeout, time.Millisecond)

	nested.release()
	active, _, ok = l.state("a")
	require.True(t, ok)
	assert.Equal(t, 1, active)
	select {
	case <-foreign:
		t.Fatal("foreign call entered before the outer call finished")
	default:
	}

	owner.release()
	got := recvWithin(t, foreign)
	require.NotNil(t, got)
	got.release()
	assert.Zero(t, l.len())
}

func TestTurnLockEmptyReentrancyIDDoesNotReenter(t *testing.T) {
	t.Parallel()
	l := newTurnLock()
	owner, err := l.acquire(context.Background(), "a", "")
	require.NoError(t, err)
	owner.allowReentrancy()

	entered := make(chan struct{}, 1)
	go func() {
		lease, acquireErr := l.acquire(context.Background(), "a", "")
		if acquireErr == nil {
			lease.release()
		}
		entered <- struct{}{}
	}()
	require.Eventually(t, func() bool {
		_, waiters, found := l.state("a")
		return found && waiters == 1
	}, turnTestTimeout, time.Millisecond)
	select {
	case <-entered:
		t.Fatal("empty reentrancy id entered an active turn")
	default:
	}
	owner.release()
	recvWithin(t, entered)
	assert.Zero(t, l.len())
}

func TestTurnLockConcurrentChurn(t *testing.T) {
	t.Parallel()
	l := newTurnLock()
	var wg sync.WaitGroup
	for id := range 16 {
		actorID := strconv.Itoa(id)
		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				lease, acquireErr := l.acquire(context.Background(), actorID, "r")
				if acquireErr != nil {
					return
				}
				defer lease.release()
				lease.allowReentrancy()
				var inner sync.WaitGroup
				for range 2 {
					inner.Go(func() {
						child, err := l.acquire(context.Background(), actorID, "r")
						if err != nil {
							return
						}
						child.release()
					})
				}
				inner.Wait()
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if id%2 == 0 {
					cancel()
				}
				lease, acquireErr := l.acquire(ctx, actorID, "")
				if acquireErr != nil {
					return
				}
				lease.release()
			}()
		}
	}
	wg.Wait()
	assert.Zero(t, l.len())
}

func TestCanceledBeforeTurnDoesNotRunActor(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := m.InvokeMethod(ctx, "id", "Count", nil)
	require.Equal(t, actorErr.ErrActorTurnCanceled, err)
	assert.Zero(t, m.createdCount())
	assert.Zero(t, m.turns.len())

	assert.Equal(t, int64(1), m.count(t, context.Background(), "id"))
	assert.Zero(t, m.turns.len())
}

func TestCanceledRunningCallKeepsTurnUntilSave(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	m.SetReentrancyEnabled(true)
	ctx := actor.WithReentrancyID(context.Background(), "chain")
	require.Equal(t, int64(1), m.count(t, ctx, "id"))
	target, _ := m.live(t, "id")

	held := make(chan struct{})
	releaseMethod, methodWait := closeOnce()
	defer releaseMethod()
	target.hold = func() {
		close(held)
		<-methodWait
	}
	saveEntered := make(chan struct{})
	releaseSave, saveWait := closeOnce()
	defer releaseSave()
	target.onSave = func() {
		close(saveEntered)
		<-saveWait
	}

	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(callCtx, "id", "Hold", nil)
		done <- aerr
	}()
	<-held
	cancel()

	second := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(actor.WithActivationID(context.Background(), "b"), "id", "Count", nil)
		second <- aerr
	}()
	waitForTurnWaiter(t, m, "id")
	assertCallWaiting(t, second)
	active, _, ok := m.turns.state("id")
	require.True(t, ok)
	assert.Equal(t, 1, active)

	releaseMethod()
	<-saveEntered
	assertCallWaiting(t, second)
	active, _, ok = m.turns.state("id")
	require.True(t, ok)
	assert.Equal(t, 1, active)

	releaseSave()
	require.Equal(t, actorErr.Success, recvWithin(t, done))
	require.Equal(t, actorErr.Success, recvWithin(t, second))
	assert.Zero(t, m.turns.len())
}

func TestDifferentActorIDsRunConcurrently(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	require.Equal(t, int64(1), m.count(t, context.Background(), "one"))
	require.Equal(t, int64(1), m.count(t, context.Background(), "two"))
	first, _ := m.live(t, "one")
	second, _ := m.live(t, "two")

	started := make(chan string, 2)
	release, wait := closeOnce()
	defer release()
	first.hold = func() {
		started <- "one"
		<-wait
	}
	second.hold = func() {
		started <- "two"
		<-wait
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = m.InvokeMethod(context.Background(), "one", "Hold", nil)
	}()
	go func() {
		defer wg.Done()
		_, _ = m.InvokeMethod(context.Background(), "two", "Hold", nil)
	}()

	got := map[string]bool{}
	deadline := time.After(turnTestTimeout)
	for len(got) < 2 {
		select {
		case id := <-started:
			got[id] = true
		case <-deadline:
			t.Fatalf("actor ids did not overlap: %v", got)
		}
	}
	release()
	wg.Wait()
}

func TestReentrancyDoesNotAdmitForeignCalls(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		enabled  bool
		outerID  string
		secondID string
	}{
		{name: "disabled", enabled: false, outerID: "chain", secondID: "chain"},
		{name: "empty id", enabled: true, outerID: "chain", secondID: ""},
		{name: "mismatched id", enabled: true, outerID: "chain", secondID: "other"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newActivationTestManager(t)
			m.SetReentrancyEnabled(tc.enabled)
			outer := actor.WithReentrancyID(context.Background(), tc.outerID)
			require.Equal(t, int64(1), m.count(t, outer, "id"))
			target, _ := m.live(t, "id")
			release, wait := closeOnce()
			defer release()
			started := make(chan struct{})
			target.hold = func() {
				close(started)
				<-wait
			}
			done := make(chan actorErr.ActorErr, 1)
			go func() {
				_, aerr := m.InvokeMethod(outer, "id", "Hold", nil)
				done <- aerr
			}()
			<-started

			secondCtx := context.Background()
			if tc.secondID != "" {
				secondCtx = actor.WithReentrancyID(secondCtx, tc.secondID)
			}
			second := make(chan actorErr.ActorErr, 1)
			go func() {
				_, aerr := m.InvokeMethod(secondCtx, "id", "Count", nil)
				second <- aerr
			}()
			waitForTurnWaiter(t, m, "id")
			assertCallWaiting(t, second)
			release()
			require.Equal(t, actorErr.Success, recvWithin(t, done))
			require.Equal(t, actorErr.Success, recvWithin(t, second))
			assert.Zero(t, m.turns.len())
		})
	}
}

func TestReentrantCallOverlapsMethodButNotContainerCreation(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	m.SetReentrancyEnabled(true)
	started, resume := pauseFirstFactory(m)
	defer resume()

	ctx := actor.WithReentrancyID(context.Background(), "chain")
	first := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Count", nil)
		first <- aerr
	}()
	<-started

	second := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Count", nil)
		second <- aerr
	}()
	waitForTurnWaiter(t, m, "id")
	assertCallWaiting(t, second)
	resume()

	require.Equal(t, actorErr.Success, recvWithin(t, first))
	require.Equal(t, actorErr.Success, recvWithin(t, second))
	assert.Equal(t, 1, m.createdCount())
	target, _ := m.live(t, "id")
	assert.Equal(t, int64(2), target.count.Load())
	assert.Zero(t, m.turns.len())
}

func TestReentrantChainHoldsTurnThroughBothSaves(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	m.SetReentrancyEnabled(true)
	ctx := actor.WithReentrancyID(context.Background(), "chain")
	require.Equal(t, int64(1), m.count(t, ctx, "id"))
	target, _ := m.live(t, "id")

	releaseMethod, methodWait := closeOnce()
	defer releaseMethod()
	started := make(chan struct{})
	target.hold = func() {
		close(started)
		<-methodWait
	}
	saveEntered := make(chan struct{}, 2)
	releaseSave, saveWait := closeOnce()
	defer releaseSave()
	target.onSave = func() {
		saveEntered <- struct{}{}
		<-saveWait
	}

	outer := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Hold", nil)
		outer <- aerr
	}()
	<-started

	// Queue a foreign call before the matching one so a correct lock has to
	// let the matching id pass it, then keep the foreign call out until both
	// saves in the chain have returned.
	foreign := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(actor.WithReentrancyID(context.Background(), "other"), "id", "Count", nil)
		foreign <- aerr
	}()
	waitForTurnWaiter(t, m, "id")

	inner := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Count", nil)
		inner <- aerr
	}()
	recvWithin(t, saveEntered)
	assertCallWaiting(t, foreign)
	active, _, ok := m.turns.state("id")
	require.True(t, ok)
	assert.Equal(t, 2, active)

	releaseMethod()
	recvWithin(t, saveEntered)
	assertCallWaiting(t, foreign)
	assertCallWaiting(t, outer)
	assertCallWaiting(t, inner)

	releaseSave()
	require.Equal(t, actorErr.Success, recvWithin(t, inner))
	require.Equal(t, actorErr.Success, recvWithin(t, outer))
	require.Equal(t, actorErr.Success, recvWithin(t, foreign))
	assert.Zero(t, m.turns.len())
}

func TestReminderAndTimerReenterMatchingTurn(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	m.SetReentrancyEnabled(true)
	ctx := actor.WithReentrancyID(context.Background(), "chain")
	require.Equal(t, int64(1), m.count(t, ctx, "id"))
	target, _ := m.live(t, "id")

	started := make(chan struct{})
	release, wait := closeOnce()
	defer release()
	target.hold = func() {
		close(started)
		<-wait
	}
	reminded := make(chan struct{})
	target.remind = func() { close(reminded) }

	done := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Hold", nil)
		done <- aerr
	}()
	<-started

	reminderDone := make(chan actorErr.ActorErr, 1)
	go func() {
		params, err := json.Marshal(&api.ActorReminderParams{DueTime: "1s", Period: "1s"})
		if err != nil {
			reminderDone <- actorErr.ErrRemindersParamsInvalid
			return
		}
		reminderDone <- m.InvokeReminder(ctx, "id", "r", params)
	}()
	recvWithin(t, reminded)
	require.Equal(t, actorErr.Success, recvWithin(t, reminderDone))

	timerDone := make(chan actorErr.ActorErr, 1)
	go func() {
		params, err := json.Marshal(&api.ActorTimerParam{CallBack: "Count", DueTime: "1s", Period: "1s"})
		if err != nil {
			timerDone <- actorErr.ErrTimerParamsInvalid
			return
		}
		timerDone <- m.InvokeTimer(ctx, "id", "t", params)
	}()
	require.Equal(t, actorErr.Success, recvWithin(t, timerDone))

	release()
	require.Equal(t, actorErr.Success, recvWithin(t, done))
	assert.Zero(t, m.turns.len())
}

func TestDeactivateDoesNotReenter(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	m.SetReentrancyEnabled(true)
	ctx := actor.WithReentrancyID(context.Background(), "chain")
	require.Equal(t, int64(1), m.count(t, ctx, "id"))
	target, _ := m.live(t, "id")

	started := make(chan struct{})
	release, wait := closeOnce()
	defer release()
	target.hold = func() {
		close(started)
		<-wait
	}
	done := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Hold", nil)
		done <- aerr
	}()
	<-started

	reentered := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(ctx, "id", "Count", nil)
		reentered <- aerr
	}()
	require.Equal(t, actorErr.Success, recvWithin(t, reentered))

	deactivated := make(chan actorErr.ActorErr, 1)
	go func() {
		deactivated <- m.DeactivateActor(ctx, "id")
	}()
	waitForTurnWaiter(t, m, "id")
	assertCallWaiting(t, deactivated)

	release()
	require.Equal(t, actorErr.Success, recvWithin(t, done))
	require.Equal(t, actorErr.Success, recvWithin(t, deactivated))
	_, ok := m.activeActors.Load("id")
	assert.False(t, ok)
	assert.Zero(t, m.turns.len())
}

func TestMethodReminderTimerAndDeactivateAreExclusive(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	m.SetReentrancyEnabled(true)
	ctx := actor.WithReentrancyID(context.Background(), "chain")
	require.Equal(t, int64(1), m.count(t, context.Background(), "id"))
	target, _ := m.live(t, "id")

	entered := make(chan string, 3)
	next := make(chan struct{})
	var n atomic.Int32
	target.hold = func() {
		entered <- strconv.Itoa(int(n.Add(1)))
		<-next
	}
	// Deactivate carries the active chain id and must still wait behind the others.
	methodDone := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(context.Background(), "id", "Hold", nil)
		methodDone <- aerr
	}()
	require.Equal(t, "1", recvWithin(t, entered))

	reminderDone := make(chan actorErr.ActorErr, 1)
	timerDone := make(chan actorErr.ActorErr, 1)
	deactDone := make(chan actorErr.ActorErr, 1)
	go func() {
		params, err := json.Marshal(&api.ActorReminderParams{DueTime: "1s", Period: "1s"})
		if err != nil {
			reminderDone <- actorErr.ErrRemindersParamsInvalid
			return
		}
		reminderDone <- m.InvokeReminder(context.Background(), "id", "r", params)
	}()
	waitForTurnCount(t, m, "id", 1)
	go func() {
		params, err := json.Marshal(&api.ActorTimerParam{CallBack: "Hold", DueTime: "1s", Period: "1s"})
		if err != nil {
			timerDone <- actorErr.ErrTimerParamsInvalid
			return
		}
		timerDone <- m.InvokeTimer(context.Background(), "id", "t", params)
	}()
	waitForTurnCount(t, m, "id", 2)
	go func() {
		deactDone <- m.DeactivateActor(ctx, "id")
	}()
	waitForTurnCount(t, m, "id", 3)
	assertCallWaiting(t, reminderDone)
	assertCallWaiting(t, timerDone)
	assertCallWaiting(t, deactDone)

	next <- struct{}{}
	require.Equal(t, "2", recvWithin(t, entered))
	require.Equal(t, actorErr.Success, recvWithin(t, methodDone))
	assertCallWaiting(t, timerDone)
	assertCallWaiting(t, deactDone)

	next <- struct{}{}
	require.Equal(t, "3", recvWithin(t, entered))
	require.Equal(t, actorErr.Success, recvWithin(t, reminderDone))
	assertCallWaiting(t, deactDone)

	next <- struct{}{}
	require.Equal(t, actorErr.Success, recvWithin(t, timerDone))
	require.Equal(t, actorErr.Success, recvWithin(t, deactDone))
	_, ok := m.activeActors.Load("id")
	assert.False(t, ok)
	assert.Zero(t, m.turns.len())
}

func TestPostInvocationSaveHoldsTurn(t *testing.T) {
	t.Parallel()
	reminder, err := json.Marshal(&api.ActorReminderParams{DueTime: "1s", Period: "1s"})
	require.NoError(t, err)
	timer, err := json.Marshal(&api.ActorTimerParam{CallBack: "Count", DueTime: "1s", Period: "1s"})
	require.NoError(t, err)

	cases := []struct {
		name   string
		invoke func(m *activationTestManager) actorErr.ActorErr
	}{
		{name: "method", invoke: func(m *activationTestManager) actorErr.ActorErr {
			_, aerr := m.InvokeMethod(context.Background(), "id", "Count", nil)
			return aerr
		}},
		{name: "reminder", invoke: func(m *activationTestManager) actorErr.ActorErr {
			return m.InvokeReminder(context.Background(), "id", "r", reminder)
		}},
		{name: "timer", invoke: func(m *activationTestManager) actorErr.ActorErr {
			return m.InvokeTimer(context.Background(), "id", "t", timer)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := newActivationTestManager(t)
			require.Equal(t, int64(1), m.count(t, context.Background(), "id"))
			target, _ := m.live(t, "id")
			if tc.name == "reminder" {
				target.remind = func() {}
			}
			saveEntered := make(chan struct{})
			releaseSave, saveWait := closeOnce()
			defer releaseSave()
			target.onSave = func() {
				close(saveEntered)
				<-saveWait
			}

			done := make(chan actorErr.ActorErr, 1)
			go func() { done <- tc.invoke(m) }()
			<-saveEntered

			second := make(chan actorErr.ActorErr, 1)
			go func() {
				second <- m.DeactivateActor(actor.WithReentrancyID(context.Background(), "chain"), "id")
			}()
			waitForTurnWaiter(t, m, "id")
			assertCallWaiting(t, second)
			releaseSave()
			require.Equal(t, actorErr.Success, recvWithin(t, done))
			require.Equal(t, actorErr.Success, recvWithin(t, second))
			assert.Zero(t, m.turns.len())
		})
	}
}

func TestInvalidCallbackPayloadDoesNotTakeTurn(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	require.Equal(t, int64(1), m.count(t, context.Background(), "id"))
	target, _ := m.live(t, "id")
	started := make(chan struct{})
	release, wait := closeOnce()
	defer release()
	target.hold = func() {
		close(started)
		<-wait
	}
	done := make(chan actorErr.ActorErr, 1)
	go func() {
		_, aerr := m.InvokeMethod(context.Background(), "id", "Hold", nil)
		done <- aerr
	}()
	<-started

	assert.Equal(t, actorErr.ErrRemindersParamsInvalid, m.InvokeReminder(context.Background(), "id", "r", []byte(`"nope"`)))
	assert.Equal(t, actorErr.ErrTimerParamsInvalid, m.InvokeTimer(context.Background(), "id", "t", []byte(`"nope"`)))
	_, waiters, ok := m.turns.state("id")
	require.True(t, ok)
	assert.Zero(t, waiters)

	release()
	require.Equal(t, actorErr.Success, recvWithin(t, done))
	assert.Zero(t, m.turns.len())
}

func TestIdleTurnEntriesAreRemovedUnderChurn(t *testing.T) {
	t.Parallel()
	m := newActivationTestManager(t)
	const ids = 24
	const calls = 6
	var wg sync.WaitGroup
	errCh := make(chan actorErr.ActorErr, ids*calls)
	for id := range ids {
		for n := range calls {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx := context.Background()
				if n%4 == 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					defer cancel()
					cancel()
				}
				_, aerr := m.InvokeMethod(ctx, strconv.Itoa(id), "Count", nil)
				if aerr != actorErr.Success && aerr != actorErr.ErrActorTurnCanceled {
					errCh <- aerr
				}
			}()
		}
	}
	wg.Wait()
	close(errCh)
	for aerr := range errCh {
		t.Errorf("unexpected actor error %d", aerr)
	}
	assert.Zero(t, m.turns.len())
}

func waitForTurnCount(t *testing.T, m *activationTestManager, actorID string, want int) {
	t.Helper()
	require.Eventually(t, func() bool {
		_, waiters, ok := m.turns.state(actorID)
		return ok && waiters == want
	}, turnTestTimeout, time.Millisecond)
}

func closeOnce() (func(), <-chan struct{}) {
	ch := make(chan struct{})
	var once sync.Once
	return func() { once.Do(func() { close(ch) }) }, ch
}

func recvWithin[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(turnTestTimeout):
		t.Fatal("timed out waiting")
	}
	var zero T
	return zero
}

func assertCallWaiting[T any](t *testing.T, ch <-chan T) {
	t.Helper()
	select {
	case v := <-ch:
		t.Fatalf("call finished while the turn was still held: %v", v)
	default:
	}
}
