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
	"sync"
)

// The sidecar can stop waiting before actor code finishes. Keep local
// ownership until every invocation in the turn has completed, not until its
// request is canceled.
//
// Waiters are FIFO. A call whose context is canceled while it is still queued
// leaves the queue and does not run. Once a call has entered, canceling its
// context does not drop the lease: the caller releases it after the actor
// method and the following state save return.
//
// A matching reentrancy id may enter only after AllowReentrancy, and only
// while the turn it matches is still active. Each acquire holds a reference
// on its entry, so an idle entry is removed only when nothing is active or
// queued on it. A waiter therefore cannot block on an entry that a newer
// call has already replaced.
type turnLock struct {
	mu      sync.Mutex
	entries map[string]*turnEntry
}

type turnEntry struct {
	refs              int
	active            int
	reentrancyID      string
	reentrancyAllowed bool
	waiters           []*turnWaiter
	granted           *turnWaiter
}

type turnWaiter struct {
	reentrancyID string
	ready        chan struct{}
}

type turnLease struct {
	once     sync.Once
	lock     *turnLock
	actorID  string
	entry    *turnEntry
	released bool
}

func newTurnLock() *turnLock {
	return &turnLock{entries: make(map[string]*turnEntry)}
}

func (l *turnLock) acquire(ctx context.Context, actorID, reentrancyID string) (*turnLease, error) {
	l.mu.Lock()
	entry := l.entries[actorID]
	if entry == nil {
		entry = &turnEntry{}
		l.entries[actorID] = entry
	}
	entry.refs++

	var self *turnWaiter
	for {
		if err := ctx.Err(); err != nil {
			if entry.granted == self && self != nil {
				entry.granted = nil
				entry.grantNext()
			}
			l.releaseRef(actorID, entry)
			l.mu.Unlock()
			return nil, err
		}
		if entry.tryEnter(reentrancyID, self) {
			l.mu.Unlock()
			return &turnLease{lock: l, actorID: actorID, entry: entry}, nil
		}

		self = &turnWaiter{reentrancyID: reentrancyID, ready: make(chan struct{})}
		entry.waiters = append(entry.waiters, self)
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			l.mu.Lock()
			if entry.removeWaiter(self) {
				l.releaseRef(actorID, entry)
				l.mu.Unlock()
				return nil, ctx.Err()
			}
			// Grant or reentrancy wake already removed self. The loop either
			// takes the turn, hands a canceled grant to the next waiter, or
			// waits again.
		case <-self.ready:
			l.mu.Lock()
		}
	}
}

// tryEnter claims the turn for reentrancyID. self is the waiter that was just
// woken, or nil for a new caller. The lock is held by the caller.
func (e *turnEntry) tryEnter(reentrancyID string, self *turnWaiter) bool {
	if e.active > 0 {
		if e.reentrancyAllowed && reentrancyID != "" && reentrancyID == e.reentrancyID {
			e.active++
			return true
		}
		return false
	}
	if e.granted != nil {
		if self != e.granted {
			return false
		}
		e.granted = nil
		e.begin(reentrancyID)
		return true
	}
	// The turn is idle. A waiter already in line must go first. If the queue
	// was left with nobody granted, wake the front waiter instead of starting
	// a second turn or parking every caller on a queue nothing will drain.
	if len(e.waiters) > 0 {
		e.grantNext()
		return false
	}
	e.begin(reentrancyID)
	return true
}

func (e *turnEntry) begin(reentrancyID string) {
	e.reentrancyID = reentrancyID
	e.reentrancyAllowed = false
	e.active = 1
}

func (e *turnEntry) removeWaiter(w *turnWaiter) bool {
	for i, cur := range e.waiters {
		if cur == w {
			e.waiters = append(e.waiters[:i], e.waiters[i+1:]...)
			return true
		}
	}
	return false
}

func (e *turnEntry) grantNext() {
	if e.granted != nil || len(e.waiters) == 0 {
		return
	}
	w := e.waiters[0]
	e.waiters = e.waiters[1:]
	e.granted = w
	close(w.ready)
}

func (l *turnLock) releaseRef(actorID string, entry *turnEntry) {
	entry.refs--
	// Only the entry currently in the map may remove itself. A stale entry
	// whose refs hit zero must not delete a replacement for the same actor id.
	if entry.refs == 0 && l.entries[actorID] == entry {
		delete(l.entries, actorID)
	}
}

// allowReentrancy lets calls that carry this turn's reentrancy id enter before
// the outer call finishes. Activation must have completed first: until this
// runs, even a matching id stays queued.
func (l *turnLease) allowReentrancy() {
	l.lock.mu.Lock()
	defer l.lock.mu.Unlock()
	if l.released || l.entry.reentrancyAllowed || l.entry.reentrancyID == "" {
		return
	}
	l.entry.reentrancyAllowed = true
	kept := make([]*turnWaiter, 0, len(l.entry.waiters))
	var wake []*turnWaiter
	for _, w := range l.entry.waiters {
		if w.reentrancyID == l.entry.reentrancyID {
			wake = append(wake, w)
			continue
		}
		kept = append(kept, w)
	}
	l.entry.waiters = kept
	for _, w := range wake {
		close(w.ready)
	}
}

func (l *turnLease) release() {
	l.once.Do(func() {
		l.lock.mu.Lock()
		defer l.lock.mu.Unlock()
		l.released = true
		l.entry.active--
		if l.entry.active == 0 {
			l.entry.reentrancyID = ""
			l.entry.reentrancyAllowed = false
			l.entry.grantNext()
		}
		l.lock.releaseRef(l.actorID, l.entry)
	})
}

func (l *turnLock) state(actorID string) (active, waiters int, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, found := l.entries[actorID]
	if !found {
		return 0, 0, false
	}
	waiting := len(entry.waiters)
	if entry.granted != nil {
		waiting++
	}
	return entry.active, waiting, true
}

func (l *turnLock) len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}
