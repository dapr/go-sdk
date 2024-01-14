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
