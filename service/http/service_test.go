package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoppingUnstartedService(t *testing.T) {
	s := newService("")
	assert.NotNil(t, s)
}
