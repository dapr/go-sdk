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
	"reflect"
	"testing"
)

func TestNewActorStateChange(t *testing.T) {
	type args struct {
		stateName  string
		value      interface{}
		changeKind ChangeKind
	}
	tests := []struct {
		name string
		args args
		want *ActorStateChange
	}{
		{
			name: "init",
			args: args{
				stateName:  "testStateName",
				value:      "testValue",
				changeKind: Add,
			},
			want: &ActorStateChange{stateName: "testStateName", value: "testValue", changeKind: Add},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewActorStateChange(tt.args.stateName, tt.args.value, tt.args.changeKind); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewActorStateChange() = %v, want %v", got, tt.want)
			}
		})
	}
}
