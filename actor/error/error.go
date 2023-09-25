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

package error

type ActorErrorStatus uint8

// TODO(@laurence) the classification, handle and print log of error should be optimized.
const (
	Success                       = ActorErrorStatus(0)
	ErrActorTypeNotFound          = ActorErrorStatus(1)
	ErrRemindersParamsInvalid     = ActorErrorStatus(2)
	ErrActorMethodNoFound         = ActorErrorStatus(3)
	ErrActorInvokeFailed          = ActorErrorStatus(4)
	ErrReminderFuncUndefined      = ActorErrorStatus(5)
	ErrActorMethodSerializeFailed = ActorErrorStatus(6)
	ErrActorSerializeNoFound      = ActorErrorStatus(7)
	ErrActorIDNotFound            = ActorErrorStatus(8)
	ErrActorFactoryNotSet         = ActorErrorStatus(9)
	ErrTimerParamsInvalid         = ActorErrorStatus(10)
	ErrSaveStateFailed            = ActorErrorStatus(11)
	ErrActorServerInvalid         = ActorErrorStatus(12)
)

type ActorError struct {
	Status ActorErrorStatus
	Err    error
}

func (ae *ActorError) Error() string {
	if ae.Err != nil {
		return ae.Err.Error()
	}

	return ""
}
