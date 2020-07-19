package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoppingUnstartedService(t *testing.T) {
	t.Parallel()
	s := newService("")
	err := s.Stop()
	assert.Nil(t, err)
}
