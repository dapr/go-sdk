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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestSchedulingAlpha1(t *testing.T) {
	ctx := t.Context()

	t.Run("schedule job - valid", func(t *testing.T) {
		job := NewJob("test",
			WithJobSchedule("test"),
			WithJobData(&anypb.Any{}),
			WithJobConstantFailurePolicy(),
		)

		err := testClient.ScheduleJobAlpha1(ctx, job)

		require.NoError(t, err)
	})

	t.Run("get job - valid", func(t *testing.T) {
		expected := NewJob("name",
			WithJobSchedule("@every 10s"),
			WithJobRepeats(4),
			WithJobDueTime("10s"),
			WithJobTTL("10s"),
			WithJobConstantFailurePolicy(),
			WithJobConstantFailurePolicyMaxRetries(4),
			WithJobConstantFailurePolicyInterval(time.Second*10),
		)

		resp, err := testClient.GetJobAlpha1(ctx, "name")
		require.NoError(t, err)
		assert.Equal(t, expected, resp)
	})

	t.Run("delete job - valid", func(t *testing.T) {
		err := testClient.DeleteJobAlpha1(ctx, "name")

		require.NoError(t, err)
	})
}

func TestJobBuilder(t *testing.T) {
	t.Run("basic job creation", func(t *testing.T) {
		job := NewJob("test-job")

		assert.Equal(t, "test-job", job.Name)
		assert.Nil(t, job.Schedule)
		assert.Nil(t, job.Repeats)
		assert.Nil(t, job.DueTime)
		assert.Nil(t, job.TTL)
		assert.Nil(t, job.Data)
		assert.Nil(t, job.FailurePolicy)
	})

	t.Run("job with all options and constant failure policy", func(t *testing.T) {
		job := NewJob("full-job",
			WithJobSchedule("@every 10m"),
			WithJobRepeats(5),
			WithJobDueTime("2024-12-31T23:59:59Z"),
			WithJobTTL("2h"),
			WithJobData(&anypb.Any{TypeUrl: "test", Value: []byte("test-data")}),
			WithJobConstantFailurePolicy(),
			WithJobConstantFailurePolicyMaxRetries(3),
			WithJobConstantFailurePolicyInterval(time.Minute*2),
		)

		assert.Equal(t, "full-job", job.Name)
		assert.Equal(t, "@every 10m", *job.Schedule)
		assert.Equal(t, uint32(5), *job.Repeats)
		assert.Equal(t, "2024-12-31T23:59:59Z", *job.DueTime)
		assert.Equal(t, "2h", *job.TTL)
		assert.Equal(t, &anypb.Any{TypeUrl: "test", Value: []byte("test-data")}, job.Data)
		constantPolicy, ok := job.FailurePolicy.(*JobFailurePolicyConstant)
		require.True(t, ok)
		assert.Equal(t, uint32(3), *constantPolicy.MaxRetries)
		assert.Equal(t, time.Minute*2, *constantPolicy.Interval)
	})

	t.Run("job with all options and drop failure policy", func(t *testing.T) {
		job := NewJob("full-job",
			WithJobSchedule("@every 10m"),
			WithJobRepeats(5),
			WithJobDueTime("2024-12-31T23:59:59Z"),
			WithJobTTL("2h"),
			WithJobData(&anypb.Any{TypeUrl: "test", Value: []byte("test-data")}),
			WithJobDropFailurePolicy(),
		)

		assert.Equal(t, "full-job", job.Name)
		assert.Equal(t, "@every 10m", *job.Schedule)
		assert.Equal(t, uint32(5), *job.Repeats)
		assert.Equal(t, "2024-12-31T23:59:59Z", *job.DueTime)
		assert.Equal(t, "2h", *job.TTL)
		assert.Equal(t, &anypb.Any{TypeUrl: "test", Value: []byte("test-data")}, job.Data)
		_, ok := job.FailurePolicy.(*JobFailurePolicyDrop)
		require.True(t, ok)
	})
}
