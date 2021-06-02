package grpc

import (
	"context"

	"github.com/dapr/go-sdk/client"
	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OnConfigurationEvent fired whenever configuration is updated.
func (s *Server) OnConfigurationEvent(ctx context.Context, in *pb.ConfigurationEventRequest) (*emptypb.Empty, error) {
	c := client.FromGrpcConfiguration(in.Configuration)
	err := client.OnConfigurationEvent(ctx, c)

	return &emptypb.Empty{}, err
}

// Get effective configuration in app. Daprd will call this method to do inspection.
func (s *Server) GetEffectiveConfiguration(ctx context.Context, empty *emptypb.Empty) (*pb.GetEffectiveConfigurationResponse, error) {
	cs := client.CollectEffectiveConfiguration(ctx)

	configs := make([]*v1.Configuration, 0, len(cs))
	for _, c := range cs {
		configs = append(configs, client.ToGrpcConfiguration(c))
	}

	return &pb.GetEffectiveConfigurationResponse {
		Configuration: configs,
	}, nil
}
