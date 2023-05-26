/*
Copyright 2023 The Dapr Authors
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

package client

import (
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

func TestIsCloudEvent(t *testing.T) {
	testcases := []struct {
		name     string
		event    []byte
		expected bool
	}{
		{
			name:     "empty event",
			event:    []byte{},
			expected: false,
		},
		{
			name:     "event in invalid format",
			event:    []byte(`foo`),
			expected: false,
		},
		{
			name:     "event in JSON format without cloudevent fields",
			event:    []byte(`{"foo":"bar"}`),
			expected: false,
		},
		{
			name:     "event with id, source, specversion and type",
			event:    []byte(`{"id":"123","source":"source","specversion":"1.0","type":"type"}`),
			expected: true,
		},
		{
			name:     "event with missing id",
			event:    []byte(`{"source":"source","specversion":"1.0","type":"type"}`),
			expected: false,
		},
		{
			name:     "event with extra fields",
			event:    []byte(`{"id":"123","source":"source","specversion":"1.0","type":"type","foo":"bar"}`),
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isCloudEvent(tc.event)
			if actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

// Implements an io.Reader that simulates failures (after optionally reading from a stream in full)
type failingReader struct {
	// Data to return before returning an error
	data io.Reader
	l    sync.Mutex
}

func (f *failingReader) Read(p []byte) (n int, err error) {
	f.l.Lock()
	defer f.l.Unlock()

	if f.data != nil {
		n, err := f.data.Read(p)
		if err == nil {
			return n, nil
		} else if errors.Is(err, io.EOF) {
			// Do not return io.EOF as error
			// Instead, just delete the stream
			// On the next call, we will return an error
			f.data = nil
			return n, nil
		} else {
			// Should not happen
			panic(err)
		}
	}

	return 0, errors.New("simulated")
}

// Implements an io.Reader that returns data slowly
type slowReader struct {
	// Data to return
	data io.Reader
	// Interval between every byte sent
	delay time.Duration
	l     sync.Mutex
}

func (s *slowReader) Read(p []byte) (n int, err error) {
	s.l.Lock()
	defer s.l.Unlock()

	// Sleep
	time.Sleep(s.delay)

	// Read one byte at a time
	return s.data.Read(p[0:1])
}
