package workflow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/durabletask-go/backend"
	"github.com/microsoft/durabletask-go/client"
	"github.com/microsoft/durabletask-go/task"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WorkflowRuntime struct {
	tasks  *task.TaskRegistry
	client *client.TaskHubGrpcClient

	mutex  sync.Mutex // TODO: implement
	quit   chan bool
	cancel context.CancelFunc
}

type Workflow func(ctx *Context) (any, error)

type Activity func(ctx ActivityContext) (any, error)

func NewRuntime(host string, port string) (*WorkflowRuntime, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // TODO: add timeout option
	defer cancel()

	address := fmt.Sprintf("%s:%s", host, port)

	clientConn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // TODO: config
	)
	if err != nil {
		return &WorkflowRuntime{}, fmt.Errorf("failed to create runtime - grpc connection failed: %v", err)
	}

	return &WorkflowRuntime{
		tasks:  task.NewTaskRegistry(),
		client: client.NewTaskHubGrpcClient(clientConn, backend.DefaultLogger()),
		quit:   make(chan bool),
		cancel: cancel,
	}, nil
}

func getDecorator(f interface{}) (string, error) {
	if f == nil {
		return "", errors.New("nil function name")
	}

	callSplit := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), ".")

	funcName := callSplit[len(callSplit)-1]

	if funcName == "1" {
		return "", errors.New("anonymous function name")
	}

	return funcName, nil
}

func wrapWorkflow(w Workflow) task.Orchestrator {
	return func(ctx *task.OrchestrationContext) (any, error) {
		wfCtx := &Context{orchestrationContext: ctx}
		return w(wfCtx)
	}
}

func (wr *WorkflowRuntime) RegisterWorkflow(w Workflow) error {
	wrappedOrchestration := wrapWorkflow(w)

	// get decorator for workflow
	name, err := getDecorator(w)
	if err != nil {
		return fmt.Errorf("failed to get workflow decorator: %v", err)
	}

	err = wr.tasks.AddOrchestratorN(name, wrappedOrchestration)
	return err
}

func wrapActivity(a Activity) task.Activity {
	return func(ctx task.ActivityContext) (any, error) {
		aCtx := ActivityContext{ctx: ctx}

		return a(aCtx)
	}
}

func (wr *WorkflowRuntime) RegisterActivity(a Activity) error {
	wrappedActivity := wrapActivity(a)

	// get decorator for activity
	name, err := getDecorator(a)
	if err != nil {
		return fmt.Errorf("failed to get activity decorator: %v", err)
	}

	err = wr.tasks.AddActivityN(name, wrappedActivity)
	return err
}

func (wr *WorkflowRuntime) Start() error {
	// go func start
	go func() {
		err := wr.client.StartWorkItemListener(context.Background(), wr.tasks)
		if err != nil {
			log.Fatalf("failed to start work stream: %v", err)
		}
		log.Println("work item listener started")
		<-wr.quit
		log.Println("work item listener shutdown")
	}()
	return nil
}

func (wr *WorkflowRuntime) Shutdown() error {
	// cancel grpc context
	wr.cancel()
	// send close signal
	wr.quit <- true
	log.Println("work item listener shutdown signal sent")
	return nil
}
