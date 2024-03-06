package service

import (
	dapr "github.com/dapr/go-sdk/client"
)

type Server interface {
	CloseService()
	SendStart(client dapr.Client, app_id string, service string, token string, callback_service string, params string, timeout int) error
	SendStop(client dapr.Client, app_id string, service string, token string) error
	// Note: The Methods below are not expcted to be used by consumers of the service and are used by the Poller & Subscriber
	GetAllLogs(client dapr.Client, app_id string, service string)
	DeleteStateEntry(key string) error
	StoreStateEntry(key string, value []byte) error
}
