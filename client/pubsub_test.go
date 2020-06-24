package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishEvent(t *testing.T) {
	ctx := context.Background()
	err := testClient.PublishEvent(ctx, "serving", []byte("ping"))
	assert.Nil(t, err)
}
