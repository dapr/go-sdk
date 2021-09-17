package codec

import perrors "github.com/pkg/errors"

// Codec is serializer interface
type Codec interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

// CodecFactory is factory of codec
type CodecFactory func() Codec

// codecFacotryMap stores
var codecFacotryMap = make(map[string]CodecFactory)

// SetActorCodec set
func SetActorCodec(name string, f CodecFactory) {
	codecFacotryMap[name] = f
}

// GetActorCodec gets the target codec instance
func GetActorCodec(name string) (Codec, error) {
	f, ok := codecFacotryMap[name]
	if !ok {
		return nil, perrors.Errorf("no actor codec implement named %s", name)
	}
	return f(), nil
}
