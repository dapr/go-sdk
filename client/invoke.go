package client

import (
	"context"
	"encoding/json"

	v1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	anypb "github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
)

// DataContent the service invocation content
type DataContent struct {
	// Data is the input data
	Data []byte
	// ContentType is the type of the data content
	ContentType string
}

func (c *GRPCClient) invokeServiceWithRequest(ctx context.Context, req *pb.InvokeServiceRequest) (out []byte, err error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	resp, err := c.protoClient.InvokeService(c.withAuthToken(ctx), req)
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

// InvokeService invokes service without raw data ([]byte).
func (c *GRPCClient) InvokeService(ctx context.Context, serviceID, method string) (out []byte, err error) {
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
			HttpExtension: &v1.HTTPExtension{
				Verb: v1.HTTPExtension_POST,
			},
		},
	}
	return c.invokeServiceWithRequest(ctx, req)
}

// InvokeServiceWithContent invokes service without content (data + content type).
func (c *GRPCClient) InvokeServiceWithContent(ctx context.Context, serviceID, method string, content *DataContent) (out []byte, err error) {
	if serviceID == "" {
		return nil, errors.New("serviceID is required")
	}
	if method == "" {
		return nil, errors.New("method name is required")
	}
	if content == nil {
		return nil, errors.New("content required")
	}

	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method:      method,
			Data:        &anypb.Any{Value: content.Data},
			ContentType: content.ContentType,
			HttpExtension: &v1.HTTPExtension{
				Verb: v1.HTTPExtension_POST,
			},
		},
	}

	return c.invokeServiceWithRequest(ctx, req)
}

// InvokeServiceWithCustomContent invokes service with custom content (struct + content type).
func (c *GRPCClient) InvokeServiceWithCustomContent(ctx context.Context, serviceID, method string, contentType string, content interface{}) (out []byte, err error) {
	if serviceID == "" {
		return nil, errors.New("serviceID is required")
	}
	if method == "" {
		return nil, errors.New("method name is required")
	}
	if contentType == "" {
		return nil, errors.New("content type required")
	}
	if content == nil {
		return nil, errors.New("content required")
	}

	contentData, err := json.Marshal(content)

	if err != nil {
		return nil, errors.WithMessage(err, "error serializing input struct")
	}

	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method:      method,
			Data:        &anypb.Any{Value: contentData},
			ContentType: contentType,
			HttpExtension: &v1.HTTPExtension{
				Verb: v1.HTTPExtension_POST,
			},
		},
	}

	return c.invokeServiceWithRequest(ctx, req)
}
