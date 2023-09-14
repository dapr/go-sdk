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

package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewActorStateChange(t *testing.T) {
	secs5 := int64(5)

	tests := map[string]struct {
		stateName  string
		value      any
		changeKind ChangeKind
		ttl        time.Duration
		want       *ActorStateChange
	}{
		"init": {
			stateName:  "testStateName",
			value:      "testValue",
			changeKind: Add,
			ttl:        time.Second*5 + time.Millisecond*400,
			want:       &ActorStateChange{stateName: "testStateName", value: "testValue", changeKind: Add, ttlInSeconds: &secs5},
		},
		"no TTL": {
			stateName:  "testStateName",
			value:      "testValue",
			changeKind: Add,
			ttl:        0,
			want:       &ActorStateChange{stateName: "testStateName", value: "testValue", changeKind: Add, ttlInSeconds: nil},
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, NewActorStateChange(test.stateName, test.value, test.changeKind, &test.ttl))
		})
	}
}
