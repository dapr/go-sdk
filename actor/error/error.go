package error

type ActorErr uint8

const (
	Success                   = ActorErr(0)
	ErrActorTypeNotFound      = ActorErr(1)
	ErrRemindersParamsInvalid = ActorErr(2)
	ErrActorMethodNoFound     = ActorErr(3)
	ErrActorInvokeFailed      = ActorErr(4)
	ErrReminderFuncUndefined  = ActorErr(5)
	ErrActorSerializeFailed   = ActorErr(6)
	ErrActorSerializeNoFound  = ActorErr(7)
	ErrActorIDNotFound        = ActorErr(8)
	ErrActorFactoryNotSet     = ActorErr(9)
)
