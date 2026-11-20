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
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

type tripRequest struct {
	ID      string `json:"id"`
	FailCar bool   `json:"failCar"`
}

// callActivity keeps the booking and compensation sequence independently testable.
type callActivity func(name string, input any, output any, compensation bool) error

type compensation struct {
	activity    string
	reservation string
}

// bookTrip records successful bookings, then undoes them in reverse order on failure.
func bookTrip(request tripRequest, call callActivity) ([]string, error) {
	steps := []struct{ book, cancel string }{
		{"BookFlight", "CancelFlight"},
		{"BookHotel", "CancelHotel"},
		{"BookCar", "CancelCar"},
	}
	completed := make([]compensation, 0, len(steps))
	reservations := make([]string, 0, len(steps))
	for _, step := range steps {
		var reservation string
		if err := call(step.book, request, &reservation, false); err != nil {
			failures := []error{fmt.Errorf("%s: %w", step.book, err)}
			for i := len(completed) - 1; i >= 0; i-- {
				undo := completed[i]
				// Try every compensation even if another has exhausted its retries.
				if undoErr := call(undo.activity, undo.reservation, nil, true); undoErr != nil {
					failures = append(failures, fmt.Errorf("%s: %w", undo.activity, undoErr))
				}
			}
			return nil, errors.Join(failures...)
		}
		completed = append(completed, compensation{step.cancel, reservation})
		reservations = append(reservations, reservation)
	}
	return reservations, nil
}

func BookTripWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var request tripRequest
	if err := ctx.GetInput(&request); err != nil {
		return nil, err
	}
	return bookTrip(request, func(name string, input any, output any, compensation bool) error {
		options := []workflow.CallActivityOption{workflow.WithActivityInput(input)}
		if compensation {
			options = append(options, workflow.WithActivityRetryPolicy(&workflow.RetryPolicy{
				MaxAttempts:          3,
				InitialRetryInterval: time.Second,
				BackoffCoefficient:   2,
			}))
		}
		return ctx.CallActivity(name, options...).Await(output)
	})
}

// These activities simulate external booking services. Real implementations should
// use request.ID as an idempotency key and persist reservations outside the worker.
func book(ctx workflow.ActivityContext, resource string) (any, error) {
	var request tripRequest
	if err := ctx.GetInput(&request); err != nil {
		return nil, err
	}
	if resource == "car" && request.FailCar {
		return nil, errors.New("no cars available")
	}
	reservation := request.ID + "-" + resource
	fmt.Printf("Booked %s\n", reservation)
	return reservation, nil
}

func BookFlight(ctx workflow.ActivityContext) (any, error) { return book(ctx, "flight") }
func BookHotel(ctx workflow.ActivityContext) (any, error)  { return book(ctx, "hotel") }
func BookCar(ctx workflow.ActivityContext) (any, error)    { return book(ctx, "car") }

// A real cancellation must also be idempotent: retries may cancel the same booking.
func cancel(ctx workflow.ActivityContext) (any, error) {
	var reservation string
	if err := ctx.GetInput(&reservation); err != nil {
		return nil, err
	}
	fmt.Printf("Cancelled %s\n", reservation)
	return nil, nil
}

func CancelFlight(ctx workflow.ActivityContext) (any, error) { return cancel(ctx) }
func CancelHotel(ctx workflow.ActivityContext) (any, error)  { return cancel(ctx) }
func CancelCar(ctx workflow.ActivityContext) (any, error)    { return cancel(ctx) }

func run(failCar bool) error {
	registry := workflow.NewRegistry()
	if err := registry.AddWorkflow(BookTripWorkflow); err != nil {
		return err
	}
	for _, activity := range []workflow.Activity{BookFlight, BookHotel, BookCar, CancelFlight, CancelHotel, CancelCar} {
		if err := registry.AddActivity(activity); err != nil {
			return err
		}
	}
	worker, err := client.NewWorkflowClient()
	if err != nil {
		return err
	}
	ctx, stop := context.WithTimeout(context.Background(), 60*time.Second)
	defer stop()
	if err := worker.StartWorker(ctx, registry); err != nil {
		return err
	}
	id, err := worker.ScheduleWorkflow(ctx, "BookTripWorkflow", workflow.WithInput(tripRequest{ID: "trip-demo", FailCar: failCar}))
	if err != nil {
		return err
	}
	metadata, err := worker.WaitForWorkflowCompletion(ctx, id)
	if err != nil {
		return err
	}
	fmt.Printf("workflow status: %s\n", metadata.RuntimeStatus)
	expected := workflow.StatusCompleted
	if failCar {
		expected = workflow.StatusFailed
	}
	if metadata.RuntimeStatus != expected {
		return fmt.Errorf("unexpected status: %s", metadata.RuntimeStatus)
	}
	return nil
}

func main() {
	failCar := flag.Bool("fail-car", true, "simulate a car booking failure and compensate earlier bookings")
	flag.Parse()
	if err := run(*failCar); err != nil {
		log.Fatal(err)
	}
}
