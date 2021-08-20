package manager

import (
	"encoding/json"
	"github.com/dapr/go-sdk/actor/api"
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

func TestInvokeMethod(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).factory)

	data, err := mng.InvokeMethod("testActorID", "testMethodName", []byte(`"hello"`))
	assert.Nil(t, data)
	assert.Equal(t, actorErr.ErrActorFactoryNotSet, err)

	mng.RegisterActorImplFactory(mock.MockActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).factory)
	data, err = mng.InvokeMethod("testActorID", "mockMethod", []byte(`"hello"`))
	assert.Nil(t, data)
	assert.Equal(t, actorErr.ErrActorMethodNoFound, err)

	data, err = mng.InvokeMethod("testActorID", "Invoke", []byte(`"hello"`))
	assert.Equal(t, data, []byte(`"hello"`))
	assert.Equal(t, actorErr.Success, err)
}

func TestDetectiveActor(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).factory)

	err = mng.DetectiveActor("testActorID")
	assert.Equal(t, actorErr.ErrActorIDNotFound, err)

	mng.RegisterActorImplFactory(mock.MockActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).factory)
	mng.InvokeMethod("testActorID", "Invoke", []byte(`"hello"`))

	err = mng.DetectiveActor("testActorID")
	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeReminder(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).factory)

	err = mng.InvokeReminder("testActorID", "testReminderName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrActorFactoryNotSet, err)

	mng.RegisterActorImplFactory(mock.MockActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).factory)
	err = mng.InvokeReminder("testActorID", "testReminderName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrRemindersParamsInvalid, err)

	reminderParam, _ := json.Marshal(&api.ActorReminderParams{
		Data:    []byte("hello"),
		DueTime: "5s",
		Period:  "6s",
	})
	err = mng.InvokeReminder("testActorID", "testReminderName", reminderParam)
	assert.Equal(t, actorErr.Success, err)
}

func TestInvokeTimer(t *testing.T) {
	mng, err := NewDefaultActorManager("json")
	assert.NotNil(t, mng)
	assert.Equal(t, actorErr.Success, err)
	assert.Nil(t, mng.(*DefaultActorManager).factory)

	err = mng.InvokeTimer("testActorID", "testTimerName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrActorFactoryNotSet, err)

	mng.RegisterActorImplFactory(mock.MockActorImplFactory)
	assert.NotNil(t, mng.(*DefaultActorManager).factory)
	err = mng.InvokeTimer("testActorID", "testTimerName", []byte(`"hello"`))
	assert.Equal(t, actorErr.ErrTimerParamsInvalid, err)

	timerParam, _ := json.Marshal(&api.ActorTimerParam{
		Data:     []byte("hello"),
		DueTime:  "5s",
		Period:   "6s",
		CallBack: "Invoke",
	})
	err = mng.InvokeTimer("testActorID", "testTimerName", timerParam)
	assert.Equal(t, actorErr.ErrActorMethodSerializeFailed, err)

	timerParam, _ = json.Marshal(&api.ActorTimerParam{
		Data:     []byte("hello"),
		DueTime:  "5s",
		Period:   "6s",
		CallBack: "NoSuchMethod",
	})
	err = mng.InvokeTimer("testActorID", "testTimerName", timerParam)
	assert.Equal(t, actorErr.ErrActorMethodNoFound, err)

	timerParam, _ = json.Marshal(&api.ActorTimerParam{
		Data:     []byte(`"hello"`),
		DueTime:  "5s",
		Period:   "6s",
		CallBack: "Invoke",
	})
	err = mng.InvokeTimer("testActorID", "testTimerName", timerParam)
	assert.Equal(t, actorErr.Success, err)
}
