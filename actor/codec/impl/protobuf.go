package impl

import (
	"errors"
	"fmt"
	"reflect"

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
	vValue := reflect.ValueOf(v)

	if vValue.Kind() != reflect.Pointer {
		return fmt.Errorf("%w, got %T", ErrNotProtoMessage, v)
	}

	targetType := vValue.Elem().Type()

	newObj := false
	var newObjValue reflect.Value

	if targetType.Kind() == reflect.Pointer {
		newObjValue = reflect.New(targetType.Elem())
		v = newObjValue.Interface()
		newObj = true
	}

	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("%w, got %T", ErrNotProtoMessage, v)
	}

	err := proto.Unmarshal(data, m)
	if err != nil {
		return err
	}

	if newObj {
		vValue.Elem().Set(newObjValue)
	}

	return nil
}
