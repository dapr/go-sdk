package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishEvent(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx, t)
	defer closer()

	err := client.PublishEvent(ctx, "serving", []byte("ping"))
	assert.Nil(t, err)
}
