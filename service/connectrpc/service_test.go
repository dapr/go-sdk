/*
Copyright 2021 The Dapr Authors
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

package connectrpc

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoppingUnstartedService(t *testing.T) {
	s := newService("", nil)
	assert.NotNil(t, s)
	err := s.Stop()
	require.NoError(t, err)
}

func TestStoppingStartedService(t *testing.T) {
	s := newService(":3333", nil)
	assert.NotNil(t, s)

	go func() {
		if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	// Wait for the server to start
	time.Sleep(200 * time.Millisecond)
	require.NoError(t, s.Stop())
}

func TestStartingStoppedService(t *testing.T) {
	s := newService(":3333", nil)
	assert.NotNil(t, s)
	stopErr := s.Stop()
	require.NoError(t, stopErr)

	startErr := s.Start()
	require.Error(t, startErr, "expected starting a stopped server to raise an error")
	assert.Equal(t, startErr.Error(), http.ErrServerClosed.Error())
}
