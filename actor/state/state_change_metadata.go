package state

type ChangeKind string

const (
	None   = ChangeKind("")
	Add    = ChangeKind("upsert")
	Update = ChangeKind("upsert")
	Remove = ChangeKind("delete")
)

type ChangeMetadata struct {
	Kind  ChangeKind
	Value interface{}
}

func NewChangeMetadata(kind ChangeKind, value interface{}) *ChangeMetadata {
	return &ChangeMetadata{
		Kind:  kind,
		Value: value,
	}
}
