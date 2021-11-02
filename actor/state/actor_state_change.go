package state

type ActorStateChange struct {
	stateName  string
	value      interface{}
	changeKind ChangeKind
}

func NewActorStateChange(stateName string, value interface{}, changeKind ChangeKind) *ActorStateChange {
	return &ActorStateChange{stateName: stateName, value: value, changeKind: changeKind}
}
