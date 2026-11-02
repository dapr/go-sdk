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
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const lookupTimeout = 5 * time.Second

// Older sidecars forward caller headers, so a caller could forge the activation id.
var minVersion = []int{1, 19, 0}

// Gate trusts activation ids only from a sidecar that sets them itself.
type Gate struct {
	lookup func(context.Context) (string, error)

	lock    sync.Mutex
	done    bool
	trusted bool
}

// NewGate returns a Gate that reads the sidecar runtime version with lookup.
func NewGate(lookup func(context.Context) (string, error)) *Gate {
	return &Gate{lookup: lookup}
}

// ID returns id when the sidecar is trusted, else an empty id.
func (g *Gate) ID(ctx context.Context, id string) string {
	if id == "" || !g.isTrusted(ctx) {
		return ""
	}
	return id
}

func (g *Gate) isTrusted(ctx context.Context) bool {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.done {
		return g.trusted
	}

	ctx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()
	version, err := g.lookup(ctx)
	if err != nil {
		log.Printf("failed to get the sidecar runtime version, ignoring the actor activation id: %v", err)
		return false
	}

	g.done = true
	g.trusted = Supported(version)
	return g.trusted
}

// Supported reports whether version is at least 1.19.0, ignoring any build suffix.
func Supported(version string) bool {
	core, _, _ := strings.Cut(version, "+")
	core, pre, _ := strings.Cut(core, "-")
	parts := strings.Split(core, ".")
	if len(parts) != len(minVersion) {
		return false
	}

	v := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return false
		}
		v[i] = n
	}
	// Release candidates of the minimum version came before the activation id shipped.
	if c := slices.Compare(v, minVersion); c != 0 {
		return c > 0
	}
	return pre == ""
}
