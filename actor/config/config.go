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

import "github.com/dapr/go-sdk/actor/codec/constant"

// ActorConfig is Actor's configuration struct.
type ActorConfig struct {
	SerializerType         string
	ActorIdleTimeout       string
	ActorScanInterval      string
	DrainOngingCallTimeout string
	DrainBalancedActors    bool
}

// Option is option function of ActorConfig.
type Option func(config *ActorConfig)

// WithSerializerName set serializer type of the actor as @serializerType.
func WithSerializerName(serializerType string) Option {
	return func(config *ActorConfig) {
		config.SerializerType = serializerType
	}
}

// WithActorIdleTimeout set actorIdleTimeout type of the actor as @actorIdleTimeout.
func WithActorIdleTimeout(actorIdleTimeout string) Option {
	return func(config *ActorConfig) {
		config.ActorIdleTimeout = actorIdleTimeout
	}
}

// WithActorScanInterval set actorScanInterval type of the actor as @actorScanInterval.
func WithActorScanInterval(actorScanInterval string) Option {
	return func(config *ActorConfig) {
		config.ActorScanInterval = actorScanInterval
	}
}

// WithDrainOngingCallTimeout set drainOngingCallTimeout type of the actor as @drainOngingCallTimeout.
func WithDrainOngingCallTimeout(drainOngingCallTimeout string) Option {
	return func(config *ActorConfig) {
		config.DrainOngingCallTimeout = drainOngingCallTimeout
	}
}

// WithDrainBalancedActors set drainBalancedActors type of the actor as @drainBalancedActors.
func WithDrainBalancedActors(drainBalancedActors bool) Option {
	return func(config *ActorConfig) {
		config.DrainBalancedActors = drainBalancedActors
	}
}

// GetConfigFromOptions get final ActorConfig set by @opts.
func GetConfigFromOptions(opts ...Option) *ActorConfig {
	conf := &ActorConfig{
		SerializerType: constant.DefaultSerializerType,
	}
	for _, o := range opts {
		o(conf)
	}
	return conf
}
