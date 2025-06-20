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

package client

import (
	"testing"
	"time"

	"github.com/dapr/kit/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestSchedulingAlpha1(t *testing.T) {
	ctx := t.Context()

	t.Run("schedule job - valid", func(t *testing.T) {
		schedule := "test"
		err := testClient.ScheduleJobAlpha1(ctx, &Job{
			Name:          "test",
			Schedule:      &schedule,
			Data:          &anypb.Any{},
			FailurePolicy: &JobFailurePolicyConstant{},
		})

		require.NoError(t, err)
	})

	t.Run("get job - valid", func(t *testing.T) {
		expected := &Job{
			Name:     "name",
			Schedule: ptr.Of("@every 10s"),
			Repeats:  ptr.Of(uint32(4)),
			DueTime:  ptr.Of("10s"),
			TTL:      ptr.Of("10s"),
			Data:     nil,
			FailurePolicy: &JobFailurePolicyConstant{
				MaxRetries: ptr.Of(uint32(4)),
				Interval:   ptr.Of(time.Second * 10),
			},
		}

		resp, err := testClient.GetJobAlpha1(ctx, "name")
		require.NoError(t, err)
		assert.Equal(t, expected, resp)
	})

	t.Run("delete job - valid", func(t *testing.T) {
		err := testClient.DeleteJobAlpha1(ctx, "name")

		require.NoError(t, err)
	})
}
