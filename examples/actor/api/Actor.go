package api

import "context"

type ActorImpl struct {
	GetUser func(context.Context, *User) (*User, error)
	Invoke  func(context.Context, string) (string, error)
	Get     func(context.Context) (string, error)
	Post    func(context.Context, string) error
}

func (a *ActorImpl) Type() string {
	return "testActorType"
}

type User struct {
	Name string `json:"name"`
	Age  uint32 `json:"age"`
}
