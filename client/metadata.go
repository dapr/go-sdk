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
		activeActorsCount := make([]*MetadataActiveActorsCount, len(resp.ActiveActorsCount))
		for a := range resp.ActiveActorsCount {
			activeActorsCount[a] = &MetadataActiveActorsCount{
				Type:  resp.ActiveActorsCount[a].Type,
				Count: resp.ActiveActorsCount[a].Count,
			}
		}
		registeredComponents := make([]*MetadataRegisteredComponents, len(resp.RegisteredComponents))
		for r := range resp.RegisteredComponents {
			registeredComponents[r] = &MetadataRegisteredComponents{
				Name:         resp.RegisteredComponents[r].Name,
				Type:         resp.RegisteredComponents[r].Type,
				Version:      resp.RegisteredComponents[r].Version,
				Capabilities: resp.RegisteredComponents[r].Capabilities,
			}
		}
		subscriptions := make([]*MetadataSubscription, len(resp.Subscriptions))
		for s := range resp.Subscriptions {
			rules := &PubsubSubscriptionRules{}
			for r := range resp.Subscriptions[s].Rules.Rules {
				rules.Rules = append(rules.Rules, &PubsubSubscriptionRule{
					Match: resp.Subscriptions[s].Rules.Rules[r].Match,
					Path:  resp.Subscriptions[s].Rules.Rules[r].Path,
				})
			}

			subscriptions[s] = &MetadataSubscription{
				PubsubName:      resp.Subscriptions[s].PubsubName,
				Topic:           resp.Subscriptions[s].Topic,
				Metadata:        resp.Subscriptions[s].Metadata,
				Rules:           rules,
				DeadLetterTopic: resp.Subscriptions[s].DeadLetterTopic,
			}
		}
		httpEndpoints := make([]*MetadataHTTPEndpoint, len(resp.HttpEndpoints))
		for e := range resp.HttpEndpoints {
			httpEndpoints[e] = &MetadataHTTPEndpoint{
				Name: resp.HttpEndpoints[e].Name,
			}
		}
		metadata = &GetMetadataResponse{
			ID:                   resp.Id,
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
