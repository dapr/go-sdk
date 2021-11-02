package manager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	actorErr "github.com/dapr/go-sdk/actor/error"
	actorMock "github.com/dapr/go-sdk/actor/mock"
)

const mockActorID = "mockActorID"

func TestNewDefaultContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := actorMock.NewMockServer(ctrl)
	mockCodec := actorMock.NewMockCodec(ctrl)

	mockServer.EXPECT().SetID(mockActorID)
	mockServer.EXPECT().SetStateManager(gomock.Any())
	mockServer.EXPECT().SaveState()
	mockServer.EXPECT().Type()

	newContainer, aerr := NewDefaultActorContainer(mockActorID, mockServer, mockCodec)
	assert.Equal(t, actorErr.Success, aerr)
	container, ok := newContainer.(*DefaultActorContainer)

	assert.True(t, ok)
	assert.NotNil(t, container)
	assert.NotNil(t, container.actor)
	assert.NotNil(t, container.serializer)
	assert.NotNil(t, container.methodType)
}

func TestContainerInvoke(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := actorMock.NewMockServer(ctrl)
	mockCodec := actorMock.NewMockCodec(ctrl)
	param := `"param"`

	mockServer.EXPECT().SetID(mockActorID)
	mockServer.EXPECT().SetStateManager(gomock.Any())
	mockServer.EXPECT().SaveState()
	mockServer.EXPECT().Type()

	newContainer, aerr := NewDefaultActorContainer("mockActorID", mockServer, mockCodec)
	assert.Equal(t, actorErr.Success, aerr)
	container := newContainer.(*DefaultActorContainer)

	mockServer.EXPECT().Invoke(gomock.Any(), "param").Return(param, nil)
	mockCodec.EXPECT().Unmarshal([]byte(param), gomock.Any()).SetArg(1, "param").Return(nil)

	rsp, err := container.Invoke("Invoke", []byte(param))

	assert.Equal(t, 2, len(rsp))
	assert.Equal(t, actorErr.Success, err)
	assert.Equal(t, param, rsp[0].Interface().(string))
}
