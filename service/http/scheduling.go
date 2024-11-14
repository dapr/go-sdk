package http

import (
	"errors"

	"github.com/dapr/go-sdk/service/common"
)

func (s *Server) AddJobEventHandler(name string, fn common.JobEventHandler) error {
	if name == "" {
		return errors.New("job event name required")
	}
	if fn == nil {
		return errors.New("job event handler required")
	}

	return errors.New("handling http scheduling requests has not been implemented in this sdk")
}
