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
	"context"
	"log"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	commonpb "github.com/dapr/dapr/pkg/proto/common/v1"
	runtimepb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

type FailurePolicy interface {
	GetPBFailurePolicy() *commonpb.JobFailurePolicy
}

type JobFailurePolicyConstant struct {
	maxRetries *uint32
	interval   *time.Duration
}

func (f *JobFailurePolicyConstant) GetPBFailurePolicy() *commonpb.JobFailurePolicy {
	policy := &commonpb.JobFailurePolicy{
		Policy: &commonpb.JobFailurePolicy_Constant{
			Constant: &commonpb.JobFailurePolicyConstant{},
		},
	}
	if f.maxRetries != nil {
		policy.Policy.(*commonpb.JobFailurePolicy_Constant).Constant.MaxRetries = f.maxRetries
	}
	if f.interval != nil {
		policy.Policy.(*commonpb.JobFailurePolicy_Constant).Constant.Interval = &durationpb.Duration{Seconds: int64(f.interval.Seconds())}
	}
	return policy
}

type JobFailurePolicyDrop struct {
}

func (f *JobFailurePolicyDrop) GetPBFailurePolicy() *commonpb.JobFailurePolicy {
	return &commonpb.JobFailurePolicy{
		Policy: &commonpb.JobFailurePolicy_Drop{
			Drop: &commonpb.JobFailurePolicyDrop{},
		},
	}
}

func NewFailurePolicyConstant(maxRetries *uint32, interval *time.Duration) FailurePolicy {
	return &JobFailurePolicyConstant{
		maxRetries: maxRetries,
		interval:   interval,
	}
}

func NewFailurePolicyDrop() FailurePolicy {
	return &JobFailurePolicyDrop{}
}

type Job struct {
	Name          string
	Schedule      string // Optional
	Repeats       uint32 // Optional
	DueTime       string // Optional
	TTL           string // Optional
	Data          *anypb.Any
	FailurePolicy FailurePolicy
}

// ScheduleJobAlpha1 raises and schedules a job.
func (c *GRPCClient) ScheduleJobAlpha1(ctx context.Context, job *Job) error {
	// TODO: Assert job fields are defined: Name, Data
	jobRequest := &runtimepb.Job{
		Name: job.Name,
		Data: job.Data,
	}

	if job.Schedule != "" {
		jobRequest.Schedule = &job.Schedule
	}

	if job.Repeats != 0 {
		jobRequest.Repeats = &job.Repeats
	}

	if job.DueTime != "" {
		jobRequest.DueTime = &job.DueTime
	}

	if job.TTL != "" {
		jobRequest.Ttl = &job.TTL
	}

	if job.FailurePolicy != nil {
		jobRequest.FailurePolicy = job.FailurePolicy.GetPBFailurePolicy()
	}
	_, err := c.protoClient.ScheduleJobAlpha1(ctx, &runtimepb.ScheduleJobRequest{
		Job: jobRequest,
	})
	return err
}

// GetJobAlpha1 retrieves a scheduled job.
func (c *GRPCClient) GetJobAlpha1(ctx context.Context, name string) (*Job, error) {
	// TODO: Name validation
	resp, err := c.protoClient.GetJobAlpha1(ctx, &runtimepb.GetJobRequest{
		Name: name,
	})
	log.Println(resp)
	if err != nil {
		return nil, err
	}

	var failurePolicy FailurePolicy
	switch policy := resp.GetJob().GetFailurePolicy().Policy.(type) {
	case *commonpb.JobFailurePolicy_Constant:
		interval := time.Duration(policy.Constant.Interval.GetSeconds()) * time.Second
		failurePolicy = &JobFailurePolicyConstant{
			maxRetries: policy.Constant.MaxRetries,
			interval:   &interval,
		}
	case *commonpb.JobFailurePolicy_Drop:
		failurePolicy = &JobFailurePolicyDrop{}
	}

	return &Job{
		Name:          resp.GetJob().GetName(),
		Schedule:      resp.GetJob().GetSchedule(),
		Repeats:       resp.GetJob().GetRepeats(),
		DueTime:       resp.GetJob().GetDueTime(),
		TTL:           resp.GetJob().GetTtl(),
		Data:          resp.GetJob().GetData(),
		FailurePolicy: failurePolicy,
	}, nil
}

// DeleteJobAlpha1 deletes a scheduled job.
func (c *GRPCClient) DeleteJobAlpha1(ctx context.Context, name string) error {
	// TODO: Name validation
	_, err := c.protoClient.DeleteJobAlpha1(ctx, &runtimepb.DeleteJobRequest{
		Name: name,
	})
	return err
}
