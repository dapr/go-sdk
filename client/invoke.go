package client

import (
	"context"
	"encoding/json"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

// InvokeService represents the request message for Service invocation
func (c *Client) InvokeService(ctx context.Context, serviceID, method string, in []byte) (out []byte, err error) {
	if serviceID == "" {
		return nil, errors.New("nil serviceID")
	}
	if method == "" {
		return nil, errors.New("nil method")
	}

	resp, err := c.protoClient.InvokeService(ctx, &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method: method,
			Data:   &any.Any{Value: in},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking service (%s)", serviceID)
	}

	out = resp.GetData().Value
	return
}

// InvokeServiceJSON represents the request message for Service invocation with identity parameter
func (c *Client) InvokeServiceJSON(ctx context.Context, serviceID, method string, in interface{}) (out []byte, err error) {
	b, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling content")
	}
	return c.InvokeService(ctx, serviceID, method, b)
}
