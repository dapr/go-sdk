package codec

import perrors "github.com/pkg/errors"

// Codec is serializer interface.
type Codec interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

// Factory is factory of codec.
type Factory func() Codec

// codecFactoryMap stores.
var codecFactoryMap = make(map[string]Factory)

// SetActorCodec set Actor's Codec.
func SetActorCodec(name string, f Factory) {
	codecFactoryMap[name] = f
}

// GetActorCodec gets the target codec instance.
func GetActorCodec(name string) (Codec, error) {
	f, ok := codecFactoryMap[name]
	if !ok {
		return nil, perrors.Errorf("no actor codec implement named %s", name)
	}
	return f(), nil
}
