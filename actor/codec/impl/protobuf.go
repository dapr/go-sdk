package impl

import (
	"errors"
	"fmt"

	"github.com/dapr/go-sdk/actor/codec"
	"github.com/dapr/go-sdk/actor/codec/constant"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNotProtoMessage = errors.New("not a proto.Message")
)

func init() {
	codec.SetActorCodec(constant.ProtobufSerializerType, func() codec.Codec {
		return &ProtobufCodec{}
	})
}

type ProtobufCodec struct{}

func (c *ProtobufCodec) Marshal(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("%w, got %T", ErrNotProtoMessage, v)
	}

	return proto.Marshal(m)
}

func (c *ProtobufCodec) Unmarshal(data []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("%w, got %T", ErrNotProtoMessage, v)
	}

	return proto.Unmarshal(data, m)
}
