package dapr

import (
	"errors"

	"github.com/dapr/go-sdk/dapr"
	"google.golang.org/grpc"
)

// Option is a functional form of an optional parameter.
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(interface{}, []grpc.CallOption) ([]grpc.CallOption, error)

// Metadata builds an option to add additionall metadata to the outgoing request.
// TODO: use more specific, strongly typed APIs as the metadata gets better described.
func Metadata(meta map[string]string) Option {
	return func(obj interface{}, list []grpc.CallOption) ([]grpc.CallOption, error) {
		switch t := obj.(type) {
		case *dapr.InvokeServiceEnvelope:
			t.Metadata = meta
		case *dapr.InvokeBindingEnvelope:
			t.Metadata = meta
		default:
			return list, errors.New(`dapr: invalid option`)
		}
		return list, nil
	}
}

// CallOption allows direct control of the underlying GRPC call.
func CallOption(opt grpc.CallOption) Option {
	return func(obj interface{}, list []grpc.CallOption) ([]grpc.CallOption, error) {
		return append(list, opt), nil
	}
}

func applyOptions(req interface{}, opts []Option) (options []grpc.CallOption, err error) {
	for _, opt := range opts {
		if options, err = opt(req, options); err != nil {
			return nil, err
		}
	}
	return options, nil
}
