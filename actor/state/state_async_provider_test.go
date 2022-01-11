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

	"github.com/dapr/go-sdk/actor/codec"
	"github.com/dapr/go-sdk/client"
)

func TestDaprStateAsyncProvider_Apply(t *testing.T) {
	type fields struct {
		daprClient      client.Client
		stateSerializer codec.Codec
	}
	type args struct {
		actorType string
		actorID   string
		changes   []*ActorStateChange
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "empty changes",
			args: args{
				actorType: "testActor",
				actorID:   "test-0",
				changes:   nil,
			},
			wantErr: false,
		},
		{
			name: "only readonly state changes",
			args: args{
				actorType: "testActor",
				actorID:   "test-0",
				changes: []*ActorStateChange{
					{
						stateName:  "stateName1",
						value:      "Any",
						changeKind: None,
					},
					{
						stateName:  "stateName2",
						value:      "Any",
						changeKind: None,
					},
				},
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DaprStateAsyncProvider{
				daprClient:      tt.fields.daprClient,
				stateSerializer: tt.fields.stateSerializer,
			}
			if err := d.Apply(tt.args.actorType, tt.args.actorID, tt.args.changes); (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDaprStateAsyncProvider_Contains(t *testing.T) {
	type fields struct {
		daprClient      client.Client
		stateSerializer codec.Codec
	}
	type args struct {
		actorType string
		actorID   string
		stateName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DaprStateAsyncProvider{
				daprClient:      tt.fields.daprClient,
				stateSerializer: tt.fields.stateSerializer,
			}
			got, err := d.Contains(tt.args.actorType, tt.args.actorID, tt.args.stateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Contains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Contains() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDaprStateAsyncProvider_Load(t *testing.T) {
	type fields struct {
		daprClient      client.Client
		stateSerializer codec.Codec
	}
	type args struct {
		actorType string
		actorID   string
		stateName string
		reply     interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DaprStateAsyncProvider{
				daprClient:      tt.fields.daprClient,
				stateSerializer: tt.fields.stateSerializer,
			}
			if err := d.Load(tt.args.actorType, tt.args.actorID, tt.args.stateName, tt.args.reply); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewDaprStateAsyncProvider(t *testing.T) {
	type args struct {
		daprClient client.Client
	}
	tests := []struct {
		name string
		args args
		want *DaprStateAsyncProvider
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDaprStateAsyncProvider(tt.args.daprClient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDaprStateAsyncProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
