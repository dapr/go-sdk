package impl

import (
	"encoding/json"

	"github.com/dapr/go-sdk/actor/codec"
	"github.com/dapr/go-sdk/actor/codec/constant"
)

func init() {
	codec.SetActorCodec(constant.DefaultSerializerType, func() codec.Codec {
		return &JSONCodec{}
	})
}

// JSONCodec is json impl of codec.Codec.
type JSONCodec struct{}

func (j *JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j *JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
