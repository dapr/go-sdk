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
	SerializerType string
}

// ReentrancyConfig is the actor reentrancy setting advertised to daprd and
// enforced by the local turn lock. A nil config means reentrancy is disabled.
type ReentrancyConfig struct {
	Enabled       bool `json:"enabled"`
	MaxStackDepth *int `json:"maxStackDepth,omitempty"`
}

// EntityConfig overrides reentrancy for the named actor types.
// A listed type uses Reentrancy instead of the global setting. A nil
// Reentrancy disables reentrancy for those types: daprd treats an omitted
// per-entity reentrancy block as disabled, and an entitiesConfig entry
// replaces the global reentrancy config for each type it names.
type EntityConfig struct {
	Entities   []string          `json:"entities,omitempty"`
	Reentrancy *ReentrancyConfig `json:"reentrancy,omitempty"`
}

// ReentrancyEnabled reports whether actorType may reenter, using daprd's
// precedence: the last entitiesConfig entry that names the type replaces the
// global setting, and a listed type with no reentrancy block is disabled.
func ReentrancyEnabled(actorType string, global *ReentrancyConfig, entities []EntityConfig) bool {
	matched := false
	enabled := false
	for _, entity := range entities {
		for _, name := range entity.Entities {
			if name != actorType {
				continue
			}
			matched = true
			enabled = entity.Reentrancy != nil && entity.Reentrancy.Enabled
		}
	}
	if matched {
		return enabled
	}
	return global != nil && global.Enabled
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
	for _, o := range opts {
		o(conf)
	}
	return conf
}
