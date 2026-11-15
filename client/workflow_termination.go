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

package client

import (
	"errors"
	"fmt"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/workflow"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// The field numbers are part of the durabletask protobuf protocol.
	terminateRequestReasonFieldNumber     protowire.Number = 5
	executionTerminatedReasonFieldNumber  protowire.Number = 3
)

// WithTerminateReason configures the reason recorded when a workflow is
// terminated. The runtime must support the TerminateRequest.reason protocol
// field.
//
// This option remains compatible with durabletask-go versions whose generated
// TerminateRequest does not know about the field yet: protobuf unknown fields
// are preserved when the request is serialized by the gRPC client.
func WithTerminateReason(reason string) workflow.TerminateOptions {
	return workflow.TerminateOptions(func(req *protos.TerminateRequest) error {
		if req == nil {
			return errors.New("terminate request is nil")
		}

		return setStringWrapperField(
			req.ProtoReflect(),
			terminateRequestReasonFieldNumber,
			"reason",
			reason,
		)
	})
}

// GetWorkflowTerminationReason returns the reason from the first
// ExecutionTerminated event in the workflow history. The bool is false when
// the history contains no termination event with a reason. An explicitly
// empty reason is present and therefore returns true with an empty string.
func GetWorkflowTerminationReason(history *workflow.GetInstanceHistoryResponse) (string, bool, error) {
	if history == nil {
		return "", false, errors.New("workflow history is nil")
	}

	for _, event := range history.Events {
		if event == nil {
			continue
		}

		terminated := event.GetExecutionTerminated()
		if terminated == nil {
			continue
		}

		reason, found, err := getStringWrapperField(
			terminated.ProtoReflect(),
			executionTerminatedReasonFieldNumber,
			"reason",
		)
		if err != nil {
			return "", false, fmt.Errorf(
				"decode termination reason from history event %d: %w",
				event.GetEventId(),
				err,
			)
		}

		// The first termination event is authoritative, including when it
		// does not contain a reason.
		return reason, found, nil
	}

	return "", false, nil
}

// setStringWrapperField writes a google.protobuf.StringValue field. It uses
// the generated field when available and otherwise encodes the field as
// protobuf unknown data for compatibility with older generated messages.
func setStringWrapperField(message protoreflect.Message, number protowire.Number, name protoreflect.Name, value string) error {
	if !message.IsValid() {
		return errors.New("protobuf message is invalid")
	}

	if field := message.Descriptor().Fields().ByNumber(protoreflect.FieldNumber(number)); field != nil {
		if err := validateStringWrapperField(field, number, name); err != nil {
			return err
		}
		message.Set(field, protoreflect.ValueOfMessage(wrapperspb.String(value).ProtoReflect()))
		return nil
	}

	// StringValue.value is field 1. Keep the outer field present even for an
	// empty reason so callers can distinguish empty from not provided.
	wrapper := protowire.AppendTag(nil, 1, protowire.BytesType)
	wrapper = protowire.AppendString(wrapper, value)

	unknown := append([]byte(nil), message.GetUnknown()...)
	unknown = protowire.AppendTag(unknown, number, protowire.BytesType)
	unknown = protowire.AppendBytes(unknown, wrapper)
	message.SetUnknown(unknown)

	return nil
}

// getStringWrapperField reads a google.protobuf.StringValue field from a
// generated field or, for older generated messages, from protobuf unknown
// data.
func getStringWrapperField(message protoreflect.Message, number protowire.Number, name protoreflect.Name) (string, bool, error) {
	if !message.IsValid() {
		return "", false, errors.New("protobuf message is invalid")
	}

	if field := message.Descriptor().Fields().ByNumber(protoreflect.FieldNumber(number)); field != nil {
		if err := validateStringWrapperField(field, number, name); err != nil {
			return "", false, err
		}
		if !message.Has(field) {
			return "", false, nil
		}

		wrapper, ok := message.Get(field).Message().Interface().(*wrapperspb.StringValue)
		if !ok {
			return "", false, fmt.Errorf("protocol field %d (%s) is not a StringValue", number, name)
		}
		return wrapper.GetValue(), true, nil
	}

	return getStringWrapperFieldFromUnknown(message.GetUnknown(), number)
}

func validateStringWrapperField(field protoreflect.FieldDescriptor, number protowire.Number, name protoreflect.Name) error {
	if field.Name() != name ||
		field.Kind() != protoreflect.MessageKind ||
		field.Message().FullName() != "google.protobuf.StringValue" {
		return fmt.Errorf("protocol field %d (%s) is not a google.protobuf.StringValue", number, name)
	}
	return nil
}

func getStringWrapperFieldFromUnknown(unknown []byte, number protowire.Number) (string, bool, error) {
	var (
		value string
		found bool
	)

	for len(unknown) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(unknown)
		if n < 0 {
			return "", false, protowire.ParseError(n)
		}
		unknown = unknown[n:]

		if fieldNumber != number || wireType != protowire.BytesType {
			n = protowire.ConsumeFieldValue(fieldNumber, wireType, unknown)
			if n < 0 {
				return "", false, protowire.ParseError(n)
			}
			if n == 0 {
				return "", false, errors.New("unexpected end-group marker in protobuf unknown fields")
			}
			unknown = unknown[n:]
			continue
		}

		wrapper, n := protowire.ConsumeBytes(unknown)
		if n < 0 {
			return "", false, protowire.ParseError(n)
		}
		unknown = unknown[n:]

		value, err := getStringValue(wrapper)
		if err != nil {
			return "", false, err
		}
		found = true
	}

	return value, found, nil
}

func getStringValue(wrapper []byte) (string, error) {
	var value string

	for len(wrapper) > 0 {
		fieldNumber, wireType, n := protowire.ConsumeTag(wrapper)
		if n < 0 {
			return "", protowire.ParseError(n)
		}
		wrapper = wrapper[n:]

		if fieldNumber != 1 || wireType != protowire.BytesType {
			n = protowire.ConsumeFieldValue(fieldNumber, wireType, wrapper)
			if n < 0 {
				return "", protowire.ParseError(n)
			}
			if n == 0 {
				return "", errors.New("unexpected end-group marker in string wrapper")
			}
			wrapper = wrapper[n:]
			continue
		}

		value, n = protowire.ConsumeString(wrapper)
		if n < 0 {
			return "", protowire.ParseError(n)
		}
		wrapper = wrapper[n:]
	}

	return value, nil
}
