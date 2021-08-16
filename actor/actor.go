package actor

import "sync"

type Client interface {
	Type() string
	ID() string
}

type Server interface {
	ID() string
	SetID(string)
	Type() string
}

type ReminderCallee interface {
	ReminderCall(string, []byte, string, string)
}

type Factory func() Server

type ServerImplBase struct {
	once sync.Once
	id   string
}

func (b *ServerImplBase) ID() string {
	return b.id
}
func (b *ServerImplBase) SetID(id string) {
	b.once.Do(func() {
		b.id = id
	})
}
