package error

type ActorErr uint8

// TODO(@laurence) the classification, handle and print log of error should be optimized.
const (
	Success                       = ActorErr(0)
	ErrActorTypeNotFound          = ActorErr(1)
	ErrRemindersParamsInvalid     = ActorErr(2)
	ErrActorMethodNoFound         = ActorErr(3)
	ErrActorInvokeFailed          = ActorErr(4)
	ErrReminderFuncUndefined      = ActorErr(5)
	ErrActorMethodSerializeFailed = ActorErr(6)
	ErrActorSerializeNoFound      = ActorErr(7)
	ErrActorIDNotFound            = ActorErr(8)
	ErrActorFactoryNotSet         = ActorErr(9)
	ErrTimerParamsInvalid         = ActorErr(10)
	ErrSaveStateFailed            = ActorErr(11)
	ErrActorServerInvalid         = ActorErr(12)
)
