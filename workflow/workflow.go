/*
Copyright 2024 The Dapr Authors
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
package workflow

import "time"

type Metadata struct {
	InstanceID             string          `json:"id"`
	Name                   string          `json:"name"`
	RuntimeStatus          Status          `json:"status"`
	CreatedAt              time.Time       `json:"createdAt"`
	LastUpdatedAt          time.Time       `json:"lastUpdatedAt"`
	SerializedInput        string          `json:"serializedInput"`
	SerializedOutput       string          `json:"serializedOutput"`
	SerializedCustomStatus string          `json:"serializedCustomStatus"`
	FailureDetails         *FailureDetails `json:"failureDetails"`
}

type FailureDetails struct {
	Type         string          `json:"type"`
	Message      string          `json:"message"`
	StackTrace   string          `json:"stackTrace"`
	InnerFailure *FailureDetails `json:"innerFailure"`
}
