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
	"errors"
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
	constantfp := &commonpb.JobFailurePolicyConstant{}
	if f.maxRetries != nil {
		constantfp.MaxRetries = f.maxRetries
	}
	if f.interval != nil {
		constantfp.Interval = durationpb.New(*f.interval)
	}
	return &commonpb.JobFailurePolicy{
		Policy: &commonpb.JobFailurePolicy_Constant{
			Constant: constantfp,
		},
	}
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

type Job struct {
	Name          string
	Schedule      *string
	Repeats       *uint32
	DueTime       *string
	TTL           *string
	Data          *anypb.Any
	FailurePolicy FailurePolicy
}

// ScheduleJobAlpha1 raises and schedules a job.
func (c *GRPCClient) ScheduleJobAlpha1(ctx context.Context, job *Job) error {
	if job.Name == "" {
		return errors.New("job name is required")
	}
	if job.Data == nil {
		return errors.New("job data is required")
	}

	jobRequest := &runtimepb.Job{
		Name:     job.Name,
		Data:     job.Data,
		Schedule: job.Schedule,
		Repeats:  job.Repeats,
		DueTime:  job.DueTime,
		Ttl:      job.TTL,
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
	if name == "" {
		return nil, errors.New("job name is required")
	}

	resp, err := c.protoClient.GetJobAlpha1(ctx, &runtimepb.GetJobRequest{
		Name: name,
	})
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
		Schedule:      resp.GetJob().Schedule,
		Repeats:       resp.GetJob().Repeats,
		DueTime:       resp.GetJob().DueTime,
		TTL:           resp.GetJob().Ttl,
		Data:          resp.GetJob().GetData(),
		FailurePolicy: failurePolicy,
	}, nil
}

// DeleteJobAlpha1 deletes a scheduled job.
func (c *GRPCClient) DeleteJobAlpha1(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("job name is required")
	}

	_, err := c.protoClient.DeleteJobAlpha1(ctx, &runtimepb.DeleteJobRequest{
		Name: name,
	})
	return err
}
