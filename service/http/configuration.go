package http

import "context"
import  "github.com/dapr/go-sdk/service/common"

func (s *Server) SetConfigurationUpdateEventHandler(fn func(ctx context.Context, in *common.ConfigurationUpdateEvent) error) error {
	// TODO: implement later
	return nil
}
