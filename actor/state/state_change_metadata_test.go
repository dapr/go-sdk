package state

import (
	"reflect"
	"testing"
)

func TestNewChangeMetadata(t *testing.T) {
	type args struct {
		kind  ChangeKind
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want *ChangeMetadata
	}{
		{
			name: "init",
			args: args{kind: Add, value: &ChangeMetadata{}},
			want: &ChangeMetadata{
				Kind:  Add,
				Value: &ChangeMetadata{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChangeMetadata(tt.args.kind, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewChangeMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
