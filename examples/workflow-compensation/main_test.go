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
	"errors"
	"reflect"
	"testing"
)

func TestBookingCompensation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fail     string
		expected []string
	}{
		{"success", "", []string{"BookFlight", "BookHotel", "BookCar"}},
		{"flight failure", "BookFlight", []string{"BookFlight"}},
		{"hotel failure", "BookHotel", []string{"BookFlight", "BookHotel", "CancelFlight"}},
		{"car failure", "BookCar", []string{"BookFlight", "BookHotel", "BookCar", "CancelHotel", "CancelFlight"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			booked := map[string]string{}
			failure := errors.New("booking failed")
			result, err := bookTrip(tripRequest{ID: "test"}, func(name string, input any, output any, undo bool) error {
				calls = append(calls, name)
				if name == tc.fail {
					return failure
				}
				if undo {
					if output != nil {
						t.Fatal("compensation must not require a result")
					}
					if input != booked[name] {
						t.Fatalf("wrong reservation for %s: %v", name, input)
					}
					return nil
				}
				if input.(tripRequest).ID != "test" {
					t.Fatal("booking lost trip input")
				}
				reservation := "reservation-" + name
				*output.(*string) = reservation
				booked["Cancel"+name[len("Book"):]] = reservation
				return nil
			})
			if !reflect.DeepEqual(calls, tc.expected) {
				t.Fatalf("calls = %v; want %v", calls, tc.expected)
			}
			if tc.fail == "" {
				if err != nil || len(result) != 3 {
					t.Fatalf("result = %v, error = %v", result, err)
				}
			} else if !errors.Is(err, failure) || result != nil {
				t.Fatalf("original failure must survive compensation: %v, %v", result, err)
			}
		})
	}
}

func TestCompensationContinuesAfterFailure(t *testing.T) {
	bookingError := errors.New("car unavailable")
	cancellationError := errors.New("hotel cancellation unavailable")
	var cancelled []string
	_, err := bookTrip(tripRequest{ID: "test"}, func(name string, input any, output any, undo bool) error {
		if name == "BookCar" {
			return bookingError
		}
		if undo {
			cancelled = append(cancelled, name)
			if name == "CancelHotel" {
				return cancellationError
			}
			return nil
		}
		*output.(*string) = "reservation"
		return nil
	})
	if !reflect.DeepEqual(cancelled, []string{"CancelHotel", "CancelFlight"}) {
		t.Fatalf("cancelled = %v", cancelled)
	}
	if !errors.Is(err, bookingError) || !errors.Is(err, cancellationError) {
		t.Fatalf("lost failure: %v", err)
	}
}
