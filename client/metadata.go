package client

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

type GetMetadataResponse struct {
	ID                   string
	ActiveActorsCount    []*MetadataActiveActorsCount
	RegisteredComponents []*MetadataRegisteredComponents
	ExtendedMetadata     map[string]string
	Subscriptions        []*MetadataSubscription
	HTTPEndpoints        []*MetadataHTTPEndpoint
}

type MetadataActiveActorsCount struct {
	Type  string
	Count int32
}

type MetadataRegisteredComponents struct {
	Name         string
	Type         string
	Version      string
	Capabilities []string
}

type MetadataSubscription struct {
	PubsubName      string
	Topic           string
	Metadata        map[string]string
	Rules           *PubsubSubscriptionRules
	DeadLetterTopic string
}

type PubsubSubscriptionRules struct {
	Rules []*PubsubSubscriptionRule
}

type PubsubSubscriptionRule struct {
	Match string
	Path  string
}

type MetadataHTTPEndpoint struct {
	Name string
}

// GetMetadata returns the metadata of the sidecar
func (c *GRPCClient) GetMetadata(ctx context.Context) (metadata *GetMetadataResponse, err error) {
	resp, err := c.protoClient.GetMetadata(ctx, &pb.GetMetadataRequest{})
	if err != nil {
		return nil, fmt.Errorf("error invoking service: %w", err)
	}
	if resp != nil {
		activeActorsCount := make([]*MetadataActiveActorsCount, len(resp.GetActorRuntime().GetActiveActors()))
		for i, a := range resp.GetActorRuntime().GetActiveActors() {
			activeActorsCount[i] = &MetadataActiveActorsCount{
				Type:  a.GetType(),
				Count: a.GetCount(),
			}
		}
		registeredComponents := make([]*MetadataRegisteredComponents, len(resp.GetRegisteredComponents()))
		for i, r := range resp.GetRegisteredComponents() {
			registeredComponents[i] = &MetadataRegisteredComponents{
				Name:         r.GetName(),
				Type:         r.GetType(),
				Version:      r.GetVersion(),
				Capabilities: r.GetCapabilities(),
			}
		}
		subscriptions := make([]*MetadataSubscription, len(resp.GetSubscriptions()))
		for i, s := range resp.GetSubscriptions() {
			rules := &PubsubSubscriptionRules{}
			for _, r := range s.GetRules().GetRules() {
				rules.Rules = append(rules.Rules, &PubsubSubscriptionRule{
					Match: r.GetMatch(),
					Path:  r.GetPath(),
				})
			}

			subscriptions[i] = &MetadataSubscription{
				PubsubName:      s.GetPubsubName(),
				Topic:           s.GetTopic(),
				Metadata:        s.GetMetadata(),
				Rules:           rules,
				DeadLetterTopic: s.GetDeadLetterTopic(),
			}
		}
		httpEndpoints := make([]*MetadataHTTPEndpoint, len(resp.GetHttpEndpoints()))
		for i, e := range resp.GetHttpEndpoints() {
			httpEndpoints[i] = &MetadataHTTPEndpoint{
				Name: e.GetName(),
			}
		}
		metadata = &GetMetadataResponse{
			ID:                   resp.GetId(),
			ActiveActorsCount:    activeActorsCount,
			RegisteredComponents: registeredComponents,
			ExtendedMetadata:     resp.GetExtendedMetadata(),
			Subscriptions:        subscriptions,
			HTTPEndpoints:        httpEndpoints,
		}
	}

	return metadata, nil
}

// SetMetadata sets a value in the extended metadata of the sidecar
func (c *GRPCClient) SetMetadata(ctx context.Context, key, value string) error {
	if len(key) == 0 {
		return errors.New("a key is required")
	}
	req := &pb.SetMetadataRequest{
		Key:   key,
		Value: value,
	}
	_, err := c.protoClient.SetMetadata(ctx, req)
	if err != nil {
		return fmt.Errorf("error setting metadata: %w", err)
	}
	return nil
}
