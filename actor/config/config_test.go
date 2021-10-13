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
