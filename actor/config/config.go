package config

import "github.com/dapr/go-sdk/actor/codec/constant"

// ActorConfig is Actor's configuration struct.
type ActorConfig struct {
	SerializerType string
}

// Option is option function of ActorConfig.
type Option func(config *ActorConfig)

// WithSerializerName set serializer type of the actor as @serializerType.
func WithSerializerName(serializerType string) Option {
	return func(config *ActorConfig) {
		config.SerializerType = serializerType
	}
}

// GetConfigFromOptions get final ActorConfig set by @opts.
func GetConfigFromOptions(opts ...Option) *ActorConfig {
	conf := &ActorConfig{
		SerializerType: constant.DefaultSerializerType,
	}
	for _, opt := range opts {
		opt(conf)
	}
	return conf
}
