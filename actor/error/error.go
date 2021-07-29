package error

type ActorError uint8

const (
	Success                     = ActorError(0)
	ErrorActorTypeNotFound      = ActorError(1)
	ErrorActorIDNotFound        = ActorError(2) // todo remove this error
	ErrorRemindersParamsInvalid = ActorError(3)
	ErrorActorInvokeFailed      = ActorError(4)
)
