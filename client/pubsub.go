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
	"encoding/json"
	"errors"
	"fmt"
	"log"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/google/uuid"
)

const (
	rawPayload = "rawPayload"
	trueValue  = "true"
)

// PublishEventOption is the type for the functional option.
type PublishEventOption func(*pb.PublishEventRequest)

// PublishEvent publishes data onto specific pubsub topic.
func (c *GRPCClient) PublishEvent(ctx context.Context, pubsubName, topicName string, data interface{}, opts ...PublishEventOption) error {
	if pubsubName == "" {
		return errors.New("pubsubName name required")
	}
	if topicName == "" {
		return errors.New("topic name required")
	}

	request := &pb.PublishEventRequest{
		PubsubName: pubsubName,
		Topic:      topicName,
	}
	for _, o := range opts {
		o(request)
	}

	if data != nil {
		switch d := data.(type) {
		case []byte:
			request.Data = d
		case string:
			request.Data = []byte(d)
		default:
			var err error
			request.DataContentType = "application/json"
			request.Data, err = json.Marshal(d)
			if err != nil {
				return fmt.Errorf("error serializing input struct: %w", err)
			}
		}
	}

	_, err := c.protoClient.PublishEvent(c.withAuthToken(ctx), request)
	if err != nil {
		return fmt.Errorf("error publishing event unto %s topic: %w", topicName, err)
	}

	return nil
}

// PublishEventWithContentType can be passed as option to PublishEvent to set an explicit Content-Type.
func PublishEventWithContentType(contentType string) PublishEventOption {
	return func(e *pb.PublishEventRequest) {
		e.DataContentType = contentType
	}
}

// PublishEventWithMetadata can be passed as option to PublishEvent to set metadata.
func PublishEventWithMetadata(metadata map[string]string) PublishEventOption {
	return func(e *pb.PublishEventRequest) {
		e.Metadata = metadata
	}
}

// PublishEventWithRawPayload can be passed as option to PublishEvent to set rawPayload metadata.
func PublishEventWithRawPayload() PublishEventOption {
	return func(e *pb.PublishEventRequest) {
		if e.Metadata == nil {
			e.Metadata = map[string]string{rawPayload: trueValue}
		} else {
			e.Metadata[rawPayload] = trueValue
		}
	}
}

// PublishEventfromCustomContent serializes an struct and publishes its contents as data (JSON) onto topic in specific pubsub component.
// Deprecated: This method is deprecated and will be removed in a future version of the SDK. Please use `PublishEvent` instead.
func (c *GRPCClient) PublishEventfromCustomContent(ctx context.Context, pubsubName, topicName string, data interface{}) error {
	log.Println("DEPRECATED: client.PublishEventfromCustomContent is deprecated and will be removed in a future version of the SDK. Please use `PublishEvent` instead.")

	// Perform the JSON marshaling here just in case someone passed a []byte or string as data
	enc, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error serializing input struct: %w", err)
	}

	return c.PublishEvent(ctx, pubsubName, topicName, enc, PublishEventWithContentType("application/json"))
}

// PublishEventsEvent is a type of event that can be published using PublishEvents.
type PublishEventsEvent struct {
	EntryID     string
	Data        []byte
	ContentType string
	Metadata    map[string]string
}

// PublishEventsResponse is the response type for PublishEvents.
type PublishEventsResponse struct {
	Event PublishEventsEvent
	Error error
}

// PublishEventsOption is the type for the functional option.
type PublishEventsOption func(*pb.BulkPublishRequest)

// PublishEvents publishes a slice of data onto topic in specific pubsub component and returns a slice of failed events.
func (c *GRPCClient) PublishEvents(ctx context.Context, pubsubName, topicName string, slice []interface{}, opts ...PublishEventsOption) []PublishEventsResponse {
	failedEntries := make([]PublishEventsResponse, 0)

	eventMap := map[string]PublishEventsEvent{}
	for _, data := range slice {
		event, err := createPublishEventsEvent(data)
		if err != nil {
			failedEntries = append(failedEntries, PublishEventsResponse{
				Event: event,
				Error: err,
			})
			continue
		}
		eventMap[event.EntryID] = event
	}

	if pubsubName == "" {
		failedEntries = append(failedEntries,
			publishEventsResponseFromError(valuesFromMap(eventMap), errors.New("pubsubName name required"))...)
		return failedEntries
	}
	if topicName == "" {
		failedEntries = append(failedEntries,
			publishEventsResponseFromError(valuesFromMap(eventMap), errors.New("topic name required"))...)
		return failedEntries
	}

	request := &pb.BulkPublishRequest{
		PubsubName: pubsubName,
		Topic:      topicName,
	}
	for _, o := range opts {
		o(request)
	}
	entries := make([]*pb.BulkPublishRequestEntry, 0, len(eventMap))
	for _, event := range eventMap {
		entries = append(entries, &pb.BulkPublishRequestEntry{
			EntryId:     event.EntryID,
			Event:       event.Data,
			ContentType: event.ContentType,
			Metadata:    event.Metadata,
		})
	}

	res, err := c.protoClient.BulkPublishEventAlpha1(c.withAuthToken(ctx), request)
	if err != nil {
		failedEntries = append(failedEntries,
			publishEventsResponseFromError(valuesFromMap(eventMap), fmt.Errorf("error publishing events unto %s topic: %w", topicName, err))...)
		return failedEntries
	}

	for _, failedEntry := range res.FailedEntries {
		failedEntries = append(failedEntries, PublishEventsResponse{
			Event: eventMap[failedEntry.EntryId],
			Error: fmt.Errorf("error publishing event with entryID %s: %s", failedEntry.EntryId, failedEntry.Error),
		})
	}

	return failedEntries
}

// createPublishEventsEvent creates a PublishEventsEvent from an interface{}.
func createPublishEventsEvent(data interface{}) (PublishEventsEvent, error) {
	event := PublishEventsEvent{}

	switch d := data.(type) {
	case PublishEventsEvent:
		return d, nil
	case []byte:
		event.Data = d
		event.ContentType = "application/octet-stream"
	case string:
		event.Data = []byte(d)
		event.ContentType = "text/plain"
	default:
		var err error
		event.ContentType = "application/json"
		event.Data, err = json.Marshal(d)
		if err != nil {
			return PublishEventsEvent{}, fmt.Errorf("error serializing input struct: %w", err)
		}

		if isCloudEvent(event.Data) {
			event.ContentType = "application/cloudevents+json"
		}
	}

	if event.EntryID == "" {
		event.EntryID = uuid.New().String()
	}

	return event, nil
}

// publishEventsResponseFromError returns a list of PublishEventsResponse with a specific error.
func publishEventsResponseFromError(events []PublishEventsEvent, err error) []PublishEventsResponse {
	responses := make([]PublishEventsResponse, len(events))
	for i, event := range events {
		responses[i] = PublishEventsResponse{
			Event: event,
			Error: err,
		}
	}

	return responses
}

// PublishEventsWithContentType can be passed as option to PublishEvents to explicitly set the same Content-Type for all events.
func PublishEventsWithContentType(contentType string) PublishEventsOption {
	return func(r *pb.BulkPublishRequest) {
		for _, entry := range r.Entries {
			entry.ContentType = contentType
		}
	}
}

// PublishEventsWithMetadata can be passed as option to PublishEvents to set metadata.
func PublishEventsWithMetadata(metadata map[string]string) PublishEventsOption {
	return func(r *pb.BulkPublishRequest) {
		r.Metadata = metadata
	}
}

// PublishEventsWithRawPayload can be passed as option to PublishEvents to set rawPayload metadata.
func PublishEventsWithRawPayload() PublishEventsOption {
	return func(r *pb.BulkPublishRequest) {
		if r.Metadata == nil {
			r.Metadata = map[string]string{rawPayload: trueValue}
		} else {
			r.Metadata[rawPayload] = trueValue
		}
	}
}
