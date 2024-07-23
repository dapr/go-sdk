package http

import (
	"errors"
	"fmt"

	"github.com/dapr/go-sdk/service/common"
)

func (s *Server) AddJobEventHandler(name string, fn common.JobEventHandler) error {
	if name == "" {
		return fmt.Errorf("job event name required")
	}
	if fn == nil {
		return fmt.Errorf("job event handler required")
	}

	return errors.New("handling http scheduling requests has not been implemented in this sdk")
}
