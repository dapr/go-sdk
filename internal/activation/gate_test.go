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

package activation

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupported(t *testing.T) {
	t.Parallel()

	for version, want := range map[string]bool{
		"1.19.0":       true,
		"1.19.1":       true,
		"1.19.1-rc.1":  true,
		"1.19.2":       true,
		"1.20.0":       true,
		"2.0.0":        true,
		"1.19.0+build": true,
		"1.19.0-rc.2":  false,
		"1.19.0-rc.7":  false,
		"1.18.9":       false,
		"0.99.99":      false,
		"edge":         false,
		"dev":          false,
		"":             false,
		"1.19":         false,
		"1.19.1.0":     false,
		"1.x.1":        false,
		"v1.19.1":      false,
	} {
		assert.Equal(t, want, Supported(version), "version %q", version)
	}
}

func TestGate(t *testing.T) {
	t.Parallel()

	fixed := func(version string, calls *atomic.Int32) func(context.Context) (string, error) {
		return func(context.Context) (string, error) {
			calls.Add(1)
			return version, nil
		}
	}

	t.Run("new sidecar keeps the id and looks up once", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		g := NewGate(fixed("1.19.1", &calls))
		assert.Equal(t, "a", g.ID(t.Context(), "a"))
		assert.Equal(t, "b", g.ID(t.Context(), "b"))
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("old sidecar drops the id and caches it", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		g := NewGate(fixed("1.19.0-rc.3", &calls))
		assert.Empty(t, g.ID(t.Context(), "a"))
		assert.Empty(t, g.ID(t.Context(), "b"))
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("edge drops the id", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		assert.Empty(t, NewGate(fixed("edge", &calls)).ID(t.Context(), "a"))
	})

	t.Run("empty id skips the lookup", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		assert.Empty(t, NewGate(fixed("1.19.1", &calls)).ID(t.Context(), ""))
		assert.Zero(t, calls.Load())
	})

	t.Run("failed lookup drops the id then retries", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		g := NewGate(func(context.Context) (string, error) {
			if calls.Add(1) == 1 {
				return "", errors.New("unavailable")
			}
			return "1.19.1", nil
		})
		assert.Empty(t, g.ID(t.Context(), "a"))
		assert.Equal(t, "a", g.ID(t.Context(), "a"))
		assert.Equal(t, "a", g.ID(t.Context(), "a"))
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("concurrent callers share one lookup", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		release := make(chan struct{})
		g := NewGate(func(context.Context) (string, error) {
			calls.Add(1)
			<-release
			return "1.19.1", nil
		})

		var wg sync.WaitGroup
		got := make([]string, 10)
		for i := range got {
			wg.Add(1)
			go func() {
				defer wg.Done()
				got[i] = g.ID(t.Context(), "a")
			}()
		}
		close(release)
		wg.Wait()

		assert.Equal(t, int32(1), calls.Load())
		for _, id := range got {
			assert.Equal(t, "a", id)
		}
	})
}
