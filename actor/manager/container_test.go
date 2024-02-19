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

package manager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	actorErr "github.com/dapr/go-sdk/actor/error"
	actorMock "github.com/dapr/go-sdk/actor/mock"
)

const mockActorID = "mockActorID"

func TestNewDefaultContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := actorMock.NewMockServer(ctrl)
	mockServerContext := actorMock.NewMockServerContext(ctrl)
	mockCodec := actorMock.NewMockCodec(ctrl)

	mockServer.EXPECT().WithContext().Return(mockServerContext)
	mockServerContext.EXPECT().SetID(mockActorID)
	mockServerContext.EXPECT().SetStateManager(gomock.Any())
	mockServerContext.EXPECT().SaveState(gomock.Any())
	mockServerContext.EXPECT().Type()

	newContainer, err := NewDefaultActorContainer(mockActorID, mockServer, mockCodec)
	assert.Equal(t, actorErr.Success, err)
	container, ok := newContainer.(*DefaultActorContainer)

	assert.True(t, ok)
	assert.NotNil(t, container)
	assert.NotNil(t, container.actor)
	assert.NotNil(t, container.ctx.serializer)
	assert.NotNil(t, container.ctx.methodType)
}

func TestContainerInvoke(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := actorMock.NewMockServer(ctrl)
	mockServerContext := actorMock.NewMockServerContext(ctrl)
	mockCodec := actorMock.NewMockCodec(ctrl)
	param := `"param"`

	mockServer.EXPECT().WithContext().Return(mockServerContext)
	mockServerContext.EXPECT().SetID(mockActorID)
	mockServerContext.EXPECT().SetStateManager(gomock.Any())
	mockServerContext.EXPECT().SaveState(gomock.Any())
	mockServerContext.EXPECT().Type()

	newContainer, err := NewDefaultActorContainer("mockActorID", mockServer, mockCodec)
	assert.Equal(t, actorErr.Success, err)
	container := newContainer.(*DefaultActorContainer)

	mockServerContext.EXPECT().Invoke(gomock.Any(), "param").Return(param, nil)
	mockCodec.EXPECT().Unmarshal([]byte(param), gomock.Any()).SetArg(1, "param").Return(nil)

	rsp, err := container.Invoke("Invoke", []byte(param))
	require.Len(t, rsp, 2)
	require.Equal(t, actorErr.Success, err)
	assert.Equal(t, param, rsp[0].Interface().(string))
}
