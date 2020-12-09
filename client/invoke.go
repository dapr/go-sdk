package client

import (
	"context"
	"strings"

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

func verbToHTTPExtension(verb string) *v1.HTTPExtension {
	if v, ok := v1.HTTPExtension_Verb_value[strings.ToUpper(verb)]; ok {
		return &v1.HTTPExtension{Verb: v1.HTTPExtension_Verb(v)}
	}
	return &v1.HTTPExtension{Verb: v1.HTTPExtension_NONE}
}

func hasRequiredInvokeArgs(serviceID, method, verb string) error {
	if serviceID == "" {
		return errors.New("serviceID")
	}
	if method == "" {
		return errors.New("method")
	}
	if verb == "" {
		return errors.New("verb")
	}
	return nil
}

// InvokeService invokes service without raw data ([]byte).
func (c *GRPCClient) InvokeService(ctx context.Context, serviceID, method, verb string) (out []byte, err error) {
	if err := hasRequiredInvokeArgs(serviceID, method, verb); err != nil {
		return nil, errors.Wrap(err, "missing required parameter")
	}
	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method:        method,
			HttpExtension: verbToHTTPExtension(verb),
		},
	}
	return c.invokeServiceWithRequest(ctx, req)
}

// InvokeServiceWithContent invokes service without content (data + content type).
func (c *GRPCClient) InvokeServiceWithContent(ctx context.Context, serviceID, method, verb string, content *DataContent) (out []byte, err error) {
	if err := hasRequiredInvokeArgs(serviceID, method, verb); err != nil {
		return nil, errors.Wrap(err, "missing required parameter")
	}
	if content == nil {
		return nil, errors.New("content required")
	}
	req := &pb.InvokeServiceRequest{
		Id: serviceID,
		Message: &v1.InvokeRequest{
			Method:        method,
			Data:          &anypb.Any{Value: content.Data},
			ContentType:   content.ContentType,
			HttpExtension: verbToHTTPExtension(verb),
		},
	}
	return c.invokeServiceWithRequest(ctx, req)
}
