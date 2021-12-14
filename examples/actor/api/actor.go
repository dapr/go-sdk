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

package api

import "context"

type ClientStub struct {
	GetUser         func(context.Context, *User) (*User, error)
	Invoke          func(context.Context, string) (string, error)
	Get             func(context.Context) (string, error)
	Post            func(context.Context, string) error
	StartTimer      func(context.Context, *TimerRequest) error
	StopTimer       func(context.Context, *TimerRequest) error
	StartReminder   func(context.Context, *ReminderRequest) error
	StopReminder    func(context.Context, *ReminderRequest) error
	IncrementAndGet func(ctx context.Context, stateKey string) (*User, error)
}

func (a *ClientStub) Type() string {
	return "testActorType"
}

func (a *ClientStub) ID() string {
	return "ActorImplID123456"
}

type User struct {
	Name string `json:"name"`
	Age  uint32 `json:"age"`
}

type TimerRequest struct {
	TimerName string `json:"timer_name"`
	CallBack  string `json:"call_back"`
	Duration  string `json:"duration"`
	Period    string `json:"period"`
	Data      string `json:"data"`
}

type ReminderRequest struct {
	ReminderName string `json:"reminder_name"`
	Duration     string `json:"duration"`
	Period       string `json:"period"`
	Data         string `json:"data"`
}
