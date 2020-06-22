package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetSecret(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx, t)
	defer closer()

	out, err := client.GetSecret(ctx, "store", "key1", nil)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	in := make(map[string]string, 0)
	in["test"] = "value"

	out, err = client.GetSecret(ctx, "store", "key1", in)
	assert.Nil(t, err)
	assert.NotNil(t, out)
}
