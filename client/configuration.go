package client

import (
	"context"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

// ConfigurationItem represents a configuration item with key, content and other information.
type ConfigurationItem struct {
	Key      string            `json:"key"`
	Content  string            `json:"content,omitempty"`
	Group    string            `json:"group,omitempty"`
	Label    string            `json:"label,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata"`
}

// SaveConfigurationRequest is the object describing a save configuration request
type SaveConfigurationRequest struct {
	StoreName string               `json:"store_name"`
	AppID     string               `json:"appID"`
	Items     []*ConfigurationItem `json:"items"`
	Metadata  map[string]string    `json:"metadata"`
}

// GetConfigurationRequest is the object describing a get configuration request
type GetConfigurationRequest struct {
	StoreName       string            `json:"store_name"`
	AppID           string            `json:"appID"`
	Group           string            `json:"group,omitempty"`
	Label           string            `json:"group,omitempty"`
	Keys            []string          `json:"keys"`
	Metadata        map[string]string `json:"metadata"`
	SubscribeUpdate bool              `json:"subscribe_update"`
}

// DeleteConfigurationRequest is the object describing a delete configuration request
type DeleteConfigurationRequest struct {
	StoreName string            `json:"store_name"`
	AppID     string            `json:"appID"`
	Group     string            `json:"group,omitempty"`
	Label     string            `json:"group,omitempty"`
	Keys      []string          `json:"keys"`
	Metadata  map[string]string `json:"metadata"`
}

// ConfigurationUpdateEvent is the object describing a configuration update event
type ConfigurationUpdateEvent struct {
	AppID string               `json:"appID"`
	Items []*ConfigurationItem `json:"items"`
}

// GetResponse is the request object for getting configuration
type GetConfigurationResponse struct {
	Items []*ConfigurationItem `json:"items"`
}

// GetConfiguration gets configuration from configuration store.
func (c *GRPCClient) GetConfiguration(ctx context.Context, in *GetConfigurationRequest) (*GetConfigurationResponse, error) {
	if in.StoreName == "" {
		return nil, errors.New("missing required argument StoreName")
	}

	req := &pb.GetConfigurationRequest{
		StoreName:       in.StoreName,
		AppId:           in.AppID,
		Group:           in.Group,
		Label:           in.Label,
		Keys:            in.Keys,
		Metadata:        in.Metadata,
		SubscribeUpdate: in.SubscribeUpdate,
	}

	getResponse, err := c.protoClient.GetConfiguration(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting configuration")
	}

	response := GetConfigurationResponse{}
	if getResponse != nil {
		response.Items = fromGRPCConfigurationItems(getResponse.Items)
	}
	return &response, nil
}

// SaveConfiguration saves configuration into configuration store.
func (c *GRPCClient) SaveConfiguration(ctx context.Context, in *SaveConfigurationRequest) error {
	if in.StoreName == "" {
		return errors.New("missing required argument StoreName")
	}
	if len(in.Items) == 0 {
		return errors.New("missing required argument Items")
	}

	req := &pb.SaveConfigurationRequest{
		StoreName:       in.StoreName,
		AppId:           in.AppID,
		Items: toGRPCConfigurationItems(in.Items),
		Metadata:        in.Metadata,
	}

	_, err := c.protoClient.SaveConfiguration(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error saving configuration")
	}

	return nil
}

// DeleteConfiguration deletes configuration from configuration store.
func (c *GRPCClient) DeleteConfiguration(ctx context.Context, in *DeleteConfigurationRequest) error {
	if in.StoreName == "" {
		return errors.New("missing required argument StoreName")
	}

	req := &pb.DeleteConfigurationRequest{
		StoreName:       in.StoreName,
		AppId:           in.AppID,
		Group:           in.Group,
		Label:           in.Label,
		Keys:            in.Keys,
		Metadata:        in.Metadata,
	}

	_, err := c.protoClient.DeleteConfiguration(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	return nil
}

func toGRPCConfigurationItems(items []*ConfigurationItem) []*v1.ConfigurationItem {
	result := make([]*v1.ConfigurationItem, 0, len(items))

	for _, item := range items {
		result = append(result, toGRPCConfigurationItem(item))
	}

	return result
}

func toGRPCConfigurationItem(item *ConfigurationItem) *v1.ConfigurationItem {
	return &v1.ConfigurationItem{
		Key: item.Key,
		Content: item.Content,
		Group: item.Group,
		Label: item.Label,
		Tags: item.Tags,
		Metadata: item.Metadata,
	}
}

func fromGRPCConfigurationItems(items []*v1.ConfigurationItem) []*ConfigurationItem {
	result := make([]*ConfigurationItem, 0, len(items))

	for _, item := range items {
		result = append(result, fromGRPCConfigurationItem(item))
	}

	return result
}

func fromGRPCConfigurationItem(item *v1.ConfigurationItem) *ConfigurationItem {
	return &ConfigurationItem{
		Key:      item.Key,
		Content:  item.Content,
		Group:    item.Group,
		Label:    item.Label,
		Tags:     item.Tags,
		Metadata: item.Metadata,
	}
}
