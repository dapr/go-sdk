package client

import (
	"context"
	"fmt"
	"strconv"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
)

var configCache map[string]*CacheItem = map[string]*CacheItem{}
type CacheItem struct {
	Configuration *Configuration
	SubscribedHandler []*SubscribedHandler
}
type SubscribedHandler struct {
	Names []string
	Handler ConfigurationUpdateHandler
}

func saveConfigurationInCache(c *Configuration)  {
	key := fmt.Sprintf("%s||%s", c.StoreName, c.AppID)

	if cachedItem, ok := configCache[key]; ok {
		cachedItem.Configuration = c
	} else {
		configCache[key] = &CacheItem{
			Configuration: c,
			SubscribedHandler: []*SubscribedHandler{},
		}
	}
}

func saveSubscribeInCache(storeName, appID string, names []string, h ConfigurationUpdateHandler)  {
	key := fmt.Sprintf("%s||%s", storeName, appID)
	handler2save := &SubscribedHandler{
		Names:   names,
		Handler: h,
	}

	if cachedItem, ok := configCache[key]; ok {
		cachedItem.SubscribedHandler = append(cachedItem.SubscribedHandler, handler2save)
	} else {
		configCache[key] = &CacheItem{
			SubscribedHandler: []*SubscribedHandler{handler2save},
		}
	}
}

func checkSubscribeNames(storeName, appID string, newNames []string) (bool, []string) {
	key := fmt.Sprintf("%s||%s", storeName, appID)
	if cachedItem, ok := configCache[key]; ok {
		set := make(map[string]struct{})
		for _, sh := range cachedItem.SubscribedHandler {
			for _, n := range sh.Names {
				set[n] = struct{}{}
			}
		}

		len1 := len(set)
		for _, n := range newNames {
			set[n] = struct{}{}
		}
		len2 := len(set)

		if len1 == len2 {
			// no new name need to do subscribe
			return false, []string{}
		}

		keys := make([]string, 0, len(set))
		for k := range set {
			keys = append(keys, k)
		}

		return true, keys
	} else {
		return true, newNames
	}
}

// Configuration represents all the configuration items of specified application.
type Configuration struct {
	AppID     string               `json:"appID"`
	StoreName string               `json:"store_name"`
	Revision  string               `json:"revision"`
	Items     []*ConfigurationItem `json:"items"`
}

func (c *Configuration) Get(name string) string {
	for _, item := range c.Items {
		if item.Name == name {
			return item.Content
		}
	}

	return ""
}

func (c *Configuration) GetWithDefault(name string, defaultValue string) string {
	content := c.Get(name)
	if content == "" {
		content = defaultValue
	}

	return content
}

// ConfigurationItem represents a configuration item with key, content and other information.
type ConfigurationItem struct {
	Name     string            `json:"name"`
	Content  string            `json:"content,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata"`
}

// GetConfigurationRequest is the object describing a get configuration request
type GetConfigurationRequest struct {
	AppID     string            `json:"appID"`
	StoreName string            `json:"store_name"`
	Metadata  map[string]string `json:"metadata"`
}

// GetResponse is the request object for getting configuration
type GetConfigurationResponse struct {
	Items []*ConfigurationItem `json:"items"`
}

// SubscribeConfigurationRequest is the object describing a subscribe configuration request
type SubscribeConfigurationRequest struct {
	AppID     string            `json:"appID"`
	StoreName string            `json:"store_name"`
	Names     []string          `json:"names"`
	Metadata  map[string]string `json:"metadata"`
}

// Handler is the handler used to invoke the app handler
type ConfigurationUpdateHandler func(ctx context.Context, c *Configuration) error

// GetConfiguration gets configuration from configuration store.
func (c *GRPCClient) GetConfiguration(ctx context.Context, in *GetConfigurationRequest) (*Configuration, error) {
	if in.StoreName == "" {
		return nil, errors.New("missing required argument StoreName")
	}

	req := &pb.GetConfigurationRequest{
		StoreName: in.StoreName,
		AppId:     in.AppID,
		Metadata:  in.Metadata,
	}

	getResponse, err := c.protoClient.GetConfiguration(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting configuration")
	}

	response := &Configuration{
		AppID:     in.AppID,
		StoreName: in.StoreName,
		Revision:  getResponse.Configuration.Revision,
		Items:     fromGRPCConfigurationItems(getResponse.Configuration.Items),
	}
	saveConfigurationInCache(response)

	return response, nil
}

// SubscribeConfiguration subscribe configuration update event from configuration store.
func (c *GRPCClient) SubscribeConfiguration(ctx context.Context, in *SubscribeConfigurationRequest, h ConfigurationUpdateHandler) error {
	if in.StoreName == "" {
		return errors.New("missing required argument StoreName")
	}

	changed, allNames := checkSubscribeNames(in.StoreName, in.AppID, in.Names)
	if !changed {
		saveSubscribeInCache(in.StoreName, in.AppID, in.Names, h)
		return nil
	}

	req := &pb.SubscribeConfigurationRequest{
		StoreName: in.StoreName,
		AppId:     in.AppID,
		Names:      allNames,
		Metadata:  in.Metadata,
	}
	_, err := c.protoClient.SubscribeConfiguration(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrap(err, "error subscribing configuration update event")
	}

	saveSubscribeInCache(in.StoreName, in.AppID, in.Names, h)

	return nil
}

// OnConfigurationEvent fired whenever configuration is updated.
func OnConfigurationEvent(ctx context.Context, in *Configuration) error {
	key := fmt.Sprintf("%s||%s", in.StoreName, in.AppID)
	cachedItem, ok := configCache[key];
	if !ok {
		saveConfigurationInCache(in)
		return nil
	}

	// check revision, if not newer than cache, skip it
	if !isNewRevision(cachedItem.Configuration.Revision, in.Revision) {
		return nil
	}

	old := cachedItem.Configuration
	new := in
	for _, sh := range cachedItem.SubscribedHandler {
		notifyEachHandler(ctx, old, new, sh)
	}
	saveConfigurationInCache(new)

	return nil
}

func notifyEachHandler(ctx context.Context, old, new *Configuration, sh *SubscribedHandler)  {
	if isConfigurationItemChanged(ctx, old, new, sh.Names) {
		sh.Handler(ctx, new)
	}
}

func isConfigurationItemChanged(ctx context.Context, old, new *Configuration, names []string) bool {
	for _, n := range names {
		oldContent := old.Get(n)
		newContent := new.Get(n)

		if oldContent != newContent {
			return true
		}
	}
	return false
}

func isNewRevision(old, new string) bool {
	var oldRevision = convertRevision(old)
	var newRevision = convertRevision(new)
	return newRevision > oldRevision
}

func convertRevision(revision string) int {
	if result, err := strconv.Atoi(revision); err == nil {
		return result
	}
	return 0
}

func CollectEffectiveConfiguration(ctx context.Context) []*Configuration {
	r := make([]*Configuration, 0, len(configCache))
	for _, c := range configCache {
		r = append(r, &Configuration{
			AppID: c.Configuration.AppID,
			StoreName: c.Configuration.StoreName,
			Revision: c.Configuration.Revision,
			// don't return items
		})
	}
	return r
}

func FromGrpcConfiguration(in *v1.Configuration) *Configuration {
	return &Configuration{
		AppID: in.AppId,
		StoreName: in.StoreName,
		Revision: in.Revision,
		Items: fromGRPCConfigurationItems(in.Items),
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
		Name:     item.Name,
		Content:  item.Content,
		Metadata: item.Metadata,
	}
}


func ToGrpcConfiguration(in *Configuration) *v1.Configuration {
	return &v1.Configuration{
		AppId: in.AppID,
		StoreName: in.StoreName,
		Revision: in.Revision,
		Items: toGRPCConfigurationItems(in.Items),
	}
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
		Name:     item.Name,
		Content:  item.Content,
		Metadata: item.Metadata,
	}
}

