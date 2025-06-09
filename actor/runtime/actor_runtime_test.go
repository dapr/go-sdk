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

package runtime

import (
	"context"
	"testing"

	actorErr "github.com/dapr/go-sdk/actor/error"
	actorMock "github.com/dapr/go-sdk/actor/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewActorRuntime(t *testing.T) {
	rt := NewActorRuntime()
	assert.NotNil(t, rt)
}

func TestGetActorRuntime(t *testing.T) {
	rt := GetActorRuntimeInstance()
	assert.NotNil(t, rt)
}

func TestRegisterActorFactoryAndInvokeMethod(t *testing.T) {
	rt := NewActorRuntime()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := rt.InvokeActorMethod("testActorType", "mockActorID", "Invoke", []byte("param"))
	assert.Equal(t, actorErr.ErrActorTypeNotFound, err)

	mockServer := actorMock.NewMockActorManagerContext(ctrl)
	rt.ctx.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	//nolint:usetesting
	mockServer.EXPECT().InvokeMethod(context.Background(), "mockActorID", "Invoke", []byte("param")).Return([]byte("response"), actorErr.Success)
	rspData, err := rt.InvokeActorMethod("testActorType", "mockActorID", "Invoke", []byte("param"))

	assert.Equal(t, []byte("response"), rspData)
	assert.Equal(t, actorErr.Success, err)
}

func TestDeactive(t *testing.T) {
	rt := NewActorRuntime()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := rt.Deactivate("testActorType", "mockActorID")
	assert.Equal(t, actorErr.ErrActorTypeNotFound, err)

	mockServer := actorMock.NewMockActorManagerContext(ctrl)
	rt.ctx.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	mockServer.EXPECT().DeactivateActor(gomock.Any(), "mockActorID").Return(actorErr.Success)
	err = rt.Deactivate("testActorType", "mockActorID")

	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeReminder(t *testing.T) {
	rt := NewActorRuntime()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := rt.InvokeReminder("testActorType", "mockActorID", "mockReminder", []byte("param"))
	assert.Equal(t, actorErr.ErrActorTypeNotFound, err)

	mockServer := actorMock.NewMockActorManagerContext(ctrl)
	rt.ctx.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	//nolint:usetesting
	mockServer.EXPECT().InvokeReminder(context.Background(), "mockActorID", "mockReminder", []byte("param")).Return(actorErr.Success)
	err = rt.InvokeReminder("testActorType", "mockActorID", "mockReminder", []byte("param"))

	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeTimer(t *testing.T) {
	rt := NewActorRuntime()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := rt.InvokeTimer("testActorType", "mockActorID", "mockTimer", []byte("param"))
	assert.Equal(t, actorErr.ErrActorTypeNotFound, err)

	mockServer := actorMock.NewMockActorManagerContext(ctrl)
	rt.ctx.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	//nolint:usetesting
	mockServer.EXPECT().InvokeTimer(context.Background(), "mockActorID", "mockTimer", []byte("param")).Return(actorErr.Success)
	err = rt.InvokeTimer("testActorType", "mockActorID", "mockTimer", []byte("param"))

	assert.Equal(t, actorErr.Success, err)
}
