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

package mock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/dapr/go-sdk/actor"
)

// Ensure MockStateManagerContext stays in sync with the actor.StateManagerContext
// interface it mocks.
var _ actor.StateManagerContext = (*MockStateManagerContext)(nil)

func TestMockStateManagerContextSetWithTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStateManager := NewMockStateManagerContext(ctrl)

	mockStateManager.EXPECT().SetWithTTL(gomock.Any(), "stateName", "value", time.Second).Return(nil)

	err := mockStateManager.SetWithTTL(context.Background(), "stateName", "value", time.Second)
	require.NoError(t, err)
}
