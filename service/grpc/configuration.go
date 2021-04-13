package grpc

import (
	"context"
	"fmt"

	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)
import "github.com/dapr/go-sdk/service/common"

func (s *Server) SetConfigurationUpdateEventHandler(fn func(ctx context.Context, in *common.ConfigurationUpdateEvent) error) error {
	if fn == nil {
		return fmt.Errorf("configuration update handler required")
	}

	if s.configurationUpdateHandler != nil {
		return fmt.Errorf("configuration update handler aready set")
	}

	s.configurationUpdateHandler = fn

	return nil
}

func (s *Server) OnConfigurationEvent(ctx context.Context, in *pb.ConfigurationEventRequest) (*emptypb.Empty, error) {
	if s.configurationUpdateHandler == nil {
		return &emptypb.Empty{}, fmt.Errorf("configuration update handler is not set")
	}

	items := make([]*common.ConfigurationItem, 0, len(in.Items))
	for _, g := range in.Items {
		item := &common.ConfigurationItem{
			Key:      g.Key,
			Content:  g.Content,
			Group:    g.Group,
			Label:    g.Label,
			Tags:     g.Tags,
			Metadata: g.Metadata,
		}
		items = append(items, item)
	}
	event := &common.ConfigurationUpdateEvent{
		AppID: in.AppId,
		Items: items,
	}
	err := s.configurationUpdateHandler(ctx, event)

	return &emptypb.Empty{}, err
}
