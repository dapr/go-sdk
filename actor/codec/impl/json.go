package impl

import (
	"encoding/json"
	"github.com/dapr/go-sdk/actor/codec"
	"github.com/dapr/go-sdk/actor/codec/constant"
)

func init() {
	codec.SetActorCodec(constant.DefaultSerializerType, func() codec.Codec {
		return &JsonCodec{}
	})
}

// JsonCodec is json impl of codec.Codec
type JsonCodec struct {
}

func (j *JsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j *JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
