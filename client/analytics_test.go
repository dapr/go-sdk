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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticsDisabled(t *testing.T) {
	optOutVars := []string{"DO_NOT_TRACK", "SCARF_NO_ANALYTICS", "DAPR_DISABLE_ANALYTICS"}

	t.Run("enabled when no opt-out variable is set", func(t *testing.T) {
		for _, name := range optOutVars {
			t.Setenv(name, "")
		}
		assert.False(t, analyticsDisabled())
	})

	for _, name := range optOutVars {
		t.Run("disabled by "+name, func(t *testing.T) {
			for _, other := range optOutVars {
				t.Setenv(other, "")
			}
			t.Setenv(name, "1")
			assert.True(t, analyticsDisabled())
		})
	}

	t.Run("ignores falsy and unset values", func(t *testing.T) {
		for _, name := range optOutVars {
			t.Setenv(name, "")
		}
		t.Setenv("DO_NOT_TRACK", "0")
		assert.False(t, analyticsDisabled())

		t.Setenv("DO_NOT_TRACK", "false")
		assert.False(t, analyticsDisabled())
	})
}

func TestIsTruthy(t *testing.T) {
	truthy := []string{"1", "true", "TRUE", "yes", "on", " true "}
	for _, v := range truthy {
		assert.True(t, isTruthy(v), "expected %q to be truthy", v)
	}

	falsy := []string{"", "0", "false", "no", "off", "maybe"}
	for _, v := range falsy {
		assert.False(t, isTruthy(v), "expected %q to be falsy", v)
	}
}

// The reporter must be safe to call and must not block or panic, including
// when the endpoint is unreachable.
func TestReportAnalyticsDoesNotBlockOrPanic(t *testing.T) {
	t.Setenv("DAPR_DISABLE_ANALYTICS", "1")
	assert.NotPanics(t, func() {
		reportAnalytics()
		reportAnalytics()
	})
}
