package actor

type ActorImpl interface {
	ActorProxy
	ReceiveReminder(string, interface{}, string, string) []byte
	OnDeactive()
	OnActive()
}

type ActorProxy interface {
	Type() string
}

type ActorImplFactory func() ActorImpl
