package client

import (
	"context"
	"encoding/json"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	anypb "github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

// InvokeServiceWithRequest invokes service with input request
func (c *Client) InvokeServiceWithRequest(ctx context.Context, req *pb.InvokeServiceRequest) (out []byte, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	resp, err := c.protoClient.InvokeService(authContext(ctx), req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking service")
	}

	// allow for service to not return any value
	if resp != nil && resp.GetData() != nil {
		out = resp.GetData().Value
		return
	}

	out = nil
	return
}

// InvokeService invokes service without data
func (c *Client) InvokeService(ctx context.Context, serviceID, method string) (out []byte, err error) {
	if serviceID == "" {
		return nil, errors.New("nil serviceID")
	}
	if method == "" {
		return nil, errors.New("nil method")
	}
	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method: method,
		},
	}
	return c.InvokeServiceWithRequest(ctx, req)
}

// InvokeServiceWithContent invokes service without content
func (c *Client) InvokeServiceWithContent(ctx context.Context, serviceID, method, contentType string, data []byte) (out []byte, err error) {
	if serviceID == "" {
		return nil, errors.New("nil serviceID")
	}
	if method == "" {
		return nil, errors.New("nil method")
	}
	if contentType == "" {
		return nil, errors.New("nil contentType")
	}

	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method:      method,
			Data:        &anypb.Any{Value: data},
			ContentType: contentType,
		},
	}

	return c.InvokeServiceWithRequest(ctx, req)
}

// InvokeServiceJSON represents the request message for Service invocation with identity parameter
func (c *Client) InvokeServiceJSON(ctx context.Context, serviceID, method string, in interface{}) (out []byte, err error) {
	if in == nil {
		return c.InvokeService(ctx, serviceID, method)
	}
	b, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling in parameter")
	}

	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method:      method,
			Data:        &anypb.Any{Value: b},
			ContentType: "application/json",
		},
	}

	return c.InvokeServiceWithRequest(ctx, req)
}
