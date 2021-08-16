package manager

import (
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefaultActorManager(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)

	mng, err = NewDefaultActorManager("badSerializerType")
	assert.Nil(t, mng)
	assert.Equal(t, actorErr.ErrActorSerializeNoFound, err)
}

func TestRegisterActorImplFactory(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).factory)
	mng.RegisterActorImplFactory(mock.MockActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).factory)
}
