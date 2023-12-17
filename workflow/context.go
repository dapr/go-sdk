package workflow

import (
	"fmt"
	"log"
	"time"

	"github.com/microsoft/durabletask-go/task"
)

type Context struct {
	orchestrationContext *task.OrchestrationContext
}

func (wfc *Context) GetInput(v interface{}) error {
	return wfc.orchestrationContext.GetInput(&v)
}

func (wfc *Context) Name() string {
	return wfc.orchestrationContext.Name
}

func (wfc *Context) InstanceID() string {
	return fmt.Sprintf("%v", wfc.orchestrationContext.ID)
}

func (wfc *Context) CurrentUTCDateTime() time.Time {
	return wfc.orchestrationContext.CurrentTimeUtc
}

func (wfc *Context) IsReplaying() bool {
	return wfc.orchestrationContext.IsReplaying
}

func (wfc *Context) CallActivity(activity interface{}) task.Task {
	var inp string
	if err := wfc.GetInput(&inp); err != nil {
		log.Printf("unable to get activity input: %v", err)
	}
	// the call should continue despite being unable to obtain an input

	return wfc.orchestrationContext.CallActivity(activity, task.WithActivityInput(inp))
}

func (wfc *Context) CallChildWorkflow() {
	// TODO: implement
	// call suborchestrator
}

func (wfc *Context) CreateTimer() {
	// TODO: implement
}

func (wfc *Context) WaitForExternalEvent() {
	// TODO: implement
}

func (wfc *Context) ContinueAsNew() {
	// TODO: implement
}
