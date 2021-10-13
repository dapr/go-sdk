package impl

import (
	"github.com/dapr/go-sdk/actor/codec"
	"github.com/dapr/go-sdk/actor/codec/constant"

	"gopkg.in/yaml.v3"
)

func init() {
	codec.SetActorCodec(constant.YamlSerializerType, func() codec.Codec {
		return &YamlCodec{}
	})
}

// YamlCodec is json yaml of codec.Codec.
type YamlCodec struct{}

func (y *YamlCodec) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (y *YamlCodec) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
