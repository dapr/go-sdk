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
