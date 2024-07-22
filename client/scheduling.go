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

	"google.golang.org/protobuf/types/known/anypb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

type Job struct {
	Name     string
	Schedule string // Optional
	Repeats  uint32 // Optional
	DueTime  string // Optional
	TTL      string // Optional
	Data     *anypb.Any
}

// ScheduleJobAlpha1 raises and schedules a job.
func (c *GRPCClient) ScheduleJobAlpha1(ctx context.Context, job *Job) error {
	// TODO: Assert job fields are defined: Name, Data
	jobRequest := &pb.Job{
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
	_, err := c.protoClient.ScheduleJobAlpha1(ctx, &pb.ScheduleJobRequest{
		Job: jobRequest,
	})
	return err
}

// GetJobAlpha1 retrieves a scheduled job.
func (c *GRPCClient) GetJobAlpha1(ctx context.Context, name string) (*Job, error) {
	// TODO: Name validation
	resp, err := c.protoClient.GetJobAlpha1(ctx, &pb.GetJobRequest{
		Name: name,
	})
	log.Println(resp)
	if err != nil {
		return nil, err
	}
	return &Job{
		Name:     resp.GetJob().GetName(),
		Schedule: resp.GetJob().GetSchedule(),
		Repeats:  resp.GetJob().GetRepeats(),
		DueTime:  resp.GetJob().GetDueTime(),
		TTL:      resp.GetJob().GetTtl(),
		Data:     resp.GetJob().GetData(),
	}, nil
}

// DeleteJobAlpha1 deletes a scheduled job.
func (c *GRPCClient) DeleteJobAlpha1(ctx context.Context, name string) error {
	// TODO: Name validation
	_, err := c.protoClient.DeleteJobAlpha1(ctx, &pb.DeleteJobRequest{
		Name: name,
	})
	return err
}
