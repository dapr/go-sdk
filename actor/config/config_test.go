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

package config

import (
	"testing"

	"github.com/dapr/go-sdk/actor/codec/constant"

	"github.com/stretchr/testify/assert"
)

func TestRegisterActorTimer(t *testing.T) {
	t.Run("get default config without options", func(t *testing.T) {
		config := GetConfigFromOptions()
		assert.NotNil(t, config)
		assert.Equal(t, constant.DefaultSerializerType, config.SerializerType)
	})

	t.Run("get config with option", func(t *testing.T) {
		config := GetConfigFromOptions(
			WithSerializerName("mockSerializerType"),
		)
		assert.NotNil(t, config)
		assert.Equal(t, "mockSerializerType", config.SerializerType)
	})
}
