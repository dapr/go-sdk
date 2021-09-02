package actor

import (
	"sync"
)

type Client interface {
	Type() string
	ID() string
}

type Server interface {
	ID() string
	SetID(string)
	Type() string
	SetStateManager(StateManager)
	// SaveState saves the state cache of this actor instance to state store component by calling api of daprd.
	// Save state is called at two places: 1. On invocation of this actor instance. 2. When new actor starts.
	SaveState()
}

type ReminderCallee interface {
	ReminderCall(string, []byte, string, string)
}

type Factory func() Server

type ServerImplBase struct {
	stateManager StateManager
	once         sync.Once
	id           string
}

func (b *ServerImplBase) SetStateManager(mng StateManager) {
	b.stateManager = mng
}

func (b *ServerImplBase) GetStateManager() StateManager {
	return b.stateManager
}

func (b *ServerImplBase) ID() string {
	return b.id
}
func (b *ServerImplBase) SetID(id string) {
	b.once.Do(func() {
		b.id = id
	})
}

func (b *ServerImplBase) SaveState() {
	if b.stateManager != nil {
		b.stateManager.Save()
	}
}

type StateManager interface {
	Add(stateName string, value interface{}) error
	Get(stateName string, reply interface{}) error
	Set(stateName string, value interface{}) error
	Remove(stateName string) error
	Contains(stateName string) (bool, error)
	Save()
	Flush()
}
