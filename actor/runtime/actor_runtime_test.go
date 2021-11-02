package runtime

import (
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

	mockServer := actorMock.NewMockActorManager(ctrl)
	rt.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	mockServer.EXPECT().InvokeMethod("mockActorID", "Invoke", []byte("param")).Return([]byte("response"), actorErr.Success)
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

	mockServer := actorMock.NewMockActorManager(ctrl)
	rt.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	mockServer.EXPECT().DetectiveActor("mockActorID").Return(actorErr.Success)
	err = rt.Deactivate("testActorType", "mockActorID")

	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeReminder(t *testing.T) {
	rt := NewActorRuntime()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := rt.InvokeReminder("testActorType", "mockActorID", "mockReminder", []byte("param"))
	assert.Equal(t, actorErr.ErrActorTypeNotFound, err)

	mockServer := actorMock.NewMockActorManager(ctrl)
	rt.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	mockServer.EXPECT().InvokeReminder("mockActorID", "mockReminder", []byte("param")).Return(actorErr.Success)
	err = rt.InvokeReminder("testActorType", "mockActorID", "mockReminder", []byte("param"))

	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeTimer(t *testing.T) {
	rt := NewActorRuntime()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := rt.InvokeTimer("testActorType", "mockActorID", "mockTimer", []byte("param"))
	assert.Equal(t, actorErr.ErrActorTypeNotFound, err)

	mockServer := actorMock.NewMockActorManager(ctrl)
	rt.actorManagers.Store("testActorType", mockServer)

	mockServer.EXPECT().RegisterActorImplFactory(gomock.Any())
	rt.RegisterActorFactory(actorMock.ActorImplFactory)

	mockServer.EXPECT().InvokeTimer("mockActorID", "mockTimer", []byte("param")).Return(actorErr.Success)
	err = rt.InvokeTimer("testActorType", "mockActorID", "mockTimer", []byte("param"))

	assert.Equal(t, actorErr.Success, err)
}
