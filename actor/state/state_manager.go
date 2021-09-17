package state

import (
	"github.com/dapr/go-sdk/actor"
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

type ActorStateManager struct {
	ActorTypeName      string
	ActorID            string
	stateChangeTracker sync.Map // map[string]*ChangeMetadata
	stateAsyncProvider *DaprStateAsyncProvider
}

func (a *ActorStateManager) Add(stateName string, value interface{}) error {
	if stateName == "" {
		return errors.Errorf("state's name can't be empty")
	}
	exists, err := a.stateAsyncProvider.Contains(a.ActorTypeName, a.ActorID, stateName)
	if err != nil {
		return err
	}

	if val, ok := a.stateChangeTracker.Load(stateName); ok {
		metadata := val.(*ChangeMetadata)
		if metadata.Kind == Remove {
			a.stateChangeTracker.Store(stateName, &ChangeMetadata{
				Kind:  Update,
				Value: value,
			})
			return nil
		}
		return errors.Errorf("Duplicate cached state: %s", stateName)
	}
	if exists {
		return errors.Errorf("Duplicate state: %s", stateName)
	}
	a.stateChangeTracker.Store(stateName, &ChangeMetadata{
		Kind:  Add,
		Value: value,
	})
	return nil
}

func (a *ActorStateManager) Get(stateName string, reply interface{}) error {
	if stateName == "" {
		return errors.Errorf("state's name can't be empty")
	}

	if val, ok := a.stateChangeTracker.Load(stateName); ok {
		metadata := val.(*ChangeMetadata)
		if metadata.Kind == Remove {
			return errors.Errorf("state is marked for remove: %s", stateName)
		}
		replyVal := reflect.ValueOf(reply).Elem()
		metadataValue := reflect.ValueOf(metadata.Value)
		if metadataValue.Kind() == reflect.Ptr {
			replyVal.Set(metadataValue.Elem())
		} else {
			replyVal.Set(metadataValue)
		}

		return nil
	}

	err := a.stateAsyncProvider.Load(a.ActorTypeName, a.ActorID, stateName, reply)
	a.stateChangeTracker.Store(stateName, &ChangeMetadata{
		Kind:  None,
		Value: reply,
	})
	return err
}

func (a *ActorStateManager) Set(stateName string, value interface{}) error {
	if stateName == "" {
		return errors.Errorf("state's name can't be empty")
	}
	if val, ok := a.stateChangeTracker.Load(stateName); ok {
		metadata := val.(*ChangeMetadata)
		if metadata.Kind == None || metadata.Kind == Remove {
			metadata.Kind = Update
		}
		a.stateChangeTracker.Store(stateName, NewChangeMetadata(metadata.Kind, value))
		return nil
	}
	a.stateChangeTracker.Store(stateName, &ChangeMetadata{
		Kind:  Add,
		Value: value,
	})
	return nil
}

func (a *ActorStateManager) Remove(stateName string) error {
	if stateName == "" {
		return errors.Errorf("state's name can't be empty")
	}
	if val, ok := a.stateChangeTracker.Load(stateName); ok {
		metadata := val.(*ChangeMetadata)
		if metadata.Kind == Remove {
			return nil
		}
		if metadata.Kind == Add {
			a.stateChangeTracker.Delete(stateName)
			return nil
		}

		a.stateChangeTracker.Store(stateName, &ChangeMetadata{
			Kind:  Remove,
			Value: nil,
		})
		return nil
	}
	if exist, err := a.stateAsyncProvider.Contains(a.ActorTypeName, a.ActorID, stateName); err != nil && exist {
		a.stateChangeTracker.Store(stateName, &ChangeMetadata{
			Kind:  Remove,
			Value: nil,
		})
	}
	return nil
}

func (a *ActorStateManager) Contains(stateName string) (bool, error) {
	if stateName == "" {
		return false, errors.Errorf("state's name can't be empty")
	}
	if val, ok := a.stateChangeTracker.Load(stateName); ok {
		metadata := val.(*ChangeMetadata)
		if metadata.Kind == Remove {
			return false, nil
		}
		return true, nil
	}
	return a.stateAsyncProvider.Contains(a.ActorTypeName, a.ActorID, stateName)
}

func (a *ActorStateManager) Save() error {
	changes := make([]*ActorStateChange, 0)
	a.stateChangeTracker.Range(func(key, value interface{}) bool {
		stateName := key.(string)
		metadata := value.(*ChangeMetadata)
		changes = append(changes, NewActorStateChange(stateName, metadata.Value, metadata.Kind))
		return true
	})
	if err := a.stateAsyncProvider.Apply(a.ActorTypeName, a.ActorID, changes); err != nil {
		return err
	}
	a.Flush()
	return nil
}

func (a *ActorStateManager) Flush() {
	a.stateChangeTracker.Range(func(key, value interface{}) bool {
		stateName := key.(string)
		metadata := value.(*ChangeMetadata)
		if metadata.Kind == Remove {
			a.stateChangeTracker.Delete(stateName)
			return true
		}
		metadata = NewChangeMetadata(None, metadata.Value)
		a.stateChangeTracker.Store(stateName, metadata)
		return true
	})
}

func NewActorStateManager(actorTypeName string, actorID string, provider *DaprStateAsyncProvider) actor.StateManager {
	return &ActorStateManager{
		stateAsyncProvider: provider,
		ActorTypeName:      actorTypeName,
		ActorID:            actorID,
	}
}
