/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/runtime"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/actor-grpc/api"
	"github.com/dapr/kit/ptr"

	daprd "github.com/dapr/go-sdk/service/grpc"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

// testActorFactory returns a factory that shares a single Dapr client across
// all actor instances instead of opening (and leaking) a new connection per
// activation.
func testActorFactory(client dapr.Client) actor.FactoryContext {
	return func() actor.ServerContext {
		return &TestActor{
			daprClient: client,
		}
	}
}

type TestActor struct {
	actor.ServerImplBaseCtx
	daprClient dapr.Client
}

func (t *TestActor) Type() string {
	return "testActorGrpcType"
}

// user defined functions
func (t *TestActor) StopTimer(ctx context.Context, req *api.TimerRequest) error {
	return t.daprClient.UnregisterActorTimer(ctx, &dapr.UnregisterActorTimerRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.TimerName,
	})
}

func (t *TestActor) StartTimer(ctx context.Context, req *api.TimerRequest) error {
	return t.daprClient.RegisterActorTimer(ctx, &dapr.RegisterActorTimerRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.TimerName,
		DueTime:   req.Duration,
		Period:    req.Period,
		Data:      []byte(req.Data),
		CallBack:  req.CallBack,
	})
}

func (t *TestActor) StartReminder(ctx context.Context, req *api.ReminderRequest) error {
	return t.daprClient.RegisterActorReminder(ctx, &dapr.RegisterActorReminderRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.ReminderName,
		DueTime:   req.Duration,
		Period:    req.Period,
		Data:      []byte(req.Data),
		FailurePolicy: &dapr.JobFailurePolicyConstant{
			MaxRetries: nil,
			Interval:   ptr.Of(time.Second * 1),
		},
	})
}

func (t *TestActor) StopReminder(ctx context.Context, req *api.ReminderRequest) error {
	return t.daprClient.UnregisterActorReminder(ctx, &dapr.UnregisterActorReminderRequest{
		ActorType: t.Type(),
		ActorID:   t.ID(),
		Name:      req.ReminderName,
	})
}

func (t *TestActor) Invoke(ctx context.Context, req string) (string, error) {
	fmt.Println("get req = ", req)
	return req, nil
}

func (t *TestActor) GetUser(ctx context.Context, user *api.User) (*api.User, error) {
	fmt.Println("call get user req = ", user)
	return user, nil
}

func (t *TestActor) Get(context.Context) (string, error) {
	return "get result", nil
}

func (t *TestActor) Post(ctx context.Context, req string) error {
	fmt.Println("get post request = ", req)
	return nil
}

func (t *TestActor) IncrementAndGet(ctx context.Context, stateKey string) (*api.User, error) {
	stateData := api.User{}
	if exist, err := t.GetStateManager().Contains(ctx, stateKey); err != nil {
		fmt.Println("state manager call contains with key " + stateKey + "err = " + err.Error())
		return &stateData, err
	} else if exist {
		if err := t.GetStateManager().Get(ctx, stateKey, &stateData); err != nil {
			fmt.Println("state manager call get with key " + stateKey + "err = " + err.Error())
			return &stateData, err
		}
	}
	stateData.Age++
	if err := t.GetStateManager().Set(ctx, stateKey, stateData); err != nil {
		fmt.Printf("state manager set get with key %s and state data = %+v, error = %s", stateKey, stateData, err.Error())
		return &stateData, err
	}
	return &stateData, nil
}

func (t *TestActor) ReminderCall(reminderName string, state []byte, dueTime string, period string) {
	fmt.Println("receive reminder = ", reminderName, " state = ", string(state))
}

func main() {
	// The Dapr sidecar accepts the actor event stream only when its app
	// channel is gRPC (--app-protocol grpc), so serve the gRPC app callback
	// service. Actor callbacks are NOT delivered to this server: they all
	// arrive over the actor event stream below.
	s, err := daprd.NewService(":50051")
	if err != nil {
		logger.Fatalf("error creating gRPC service: %v", err)
	}
	go func() {
		if err := s.Start(); err != nil {
			logger.Fatalf("error listening: %v", err)
		}
	}()
	defer s.GracefulStop() //nolint:errcheck

	client, err := dapr.NewClient()
	if err != nil {
		logger.Fatalf("error creating Dapr client: %v", err)
	}
	defer client.Close()

	// Register the hosted actor types on the actor runtime, exactly as when
	// hosting actors over HTTP. The actors share the client created above.
	rt := runtime.GetActorRuntimeInstanceContext()
	rt.RegisterActorFactory(testActorFactory(client))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Host the registered actor types over the actor event stream. The
	// sidecar may still be initializing its app channel, so retry for a
	// short while before giving up.
	stop, err := subscribeWithRetry(ctx, client, rt)
	if err != nil {
		logger.Fatalf("error subscribing to actor events: %v", err)
	}
	defer stop() //nolint:errcheck

	fmt.Println("actor event subscription started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}

func subscribeWithRetry(ctx context.Context, client dapr.Client, rt *runtime.ActorRunTimeContext) (func() error, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var err error
	for i := 0; i < 30; i++ {
		var stop func() error
		stop, err = client.SubscribeActorEvents(ctx, rt, dapr.ActorEventSubscriptionOptions{
			ActorIdleTimeout: time.Hour,
		})
		if err == nil {
			return stop, nil
		}
		logger.Printf("actor event subscription not ready, retrying: %v", err)
		// Stop retrying promptly if the caller cancels (e.g. shutdown).
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
	return nil, err
}
