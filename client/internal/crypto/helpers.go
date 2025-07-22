// Package crypto was introduced when removing the code dependency to the
// generated protos from the dapr/dapr repository.
// It contains helper methods that are meant to replace the custom changes
// introduced by the protos on that repo at:
// https://github.com/dapr/dapr/blob/db5361701c55cae8ad21be60e2d4557f98cbc741/pkg/proto/runtime/v1/dapr_additional.go.
package crypto

import (
	"google.golang.org/protobuf/proto"

	commonv1pb "github.com/dapr/go-sdk/internal/proto/dapr/proto/common/v1"
	runtimev1 "github.com/dapr/go-sdk/internal/proto/dapr/proto/runtime/v1"
)

// GetPayload will retrieve the payload from one of the given runtime proto message.
func GetPayload[T runtimev1.DecryptRequest | runtimev1.EncryptRequest | runtimev1.DecryptResponse | runtimev1.EncryptResponse](req *T) *commonv1pb.StreamPayload {
	if req == nil {
		return nil
	}

	switch r := any(req).(type) {
	case *runtimev1.EncryptRequest:
		return r.GetPayload()
	case *runtimev1.DecryptRequest:
		return r.GetPayload()
	case *runtimev1.EncryptResponse:
		return r.GetPayload()
	case *runtimev1.DecryptResponse:
		return r.GetPayload()
	}

	return nil
}

// SetPayload will set the payload for one of the given runtime proto message.
func SetPayload[T runtimev1.DecryptRequest | runtimev1.EncryptRequest | runtimev1.DecryptResponse | runtimev1.EncryptResponse](req *T, payload *commonv1pb.StreamPayload) {
	if req == nil {
		return
	}

	switch r := any(req).(type) {
	case *runtimev1.EncryptRequest:
		r.Payload = payload
	case *runtimev1.DecryptRequest:
		r.Payload = payload
	case *runtimev1.EncryptResponse:
		r.Payload = payload
	case *runtimev1.DecryptResponse:
		r.Payload = payload
	}
}

// SetOptions will set the options for one of the given runtime proto message.
func SetOptions[T runtimev1.DecryptRequest | runtimev1.EncryptRequest](req *T, opts proto.Message) {
	if req == nil {
		return
	}

	switch r := any(req).(type) {
	case *runtimev1.EncryptRequest:
		r.Options = opts.(*runtimev1.EncryptRequestOptions)
	case *runtimev1.DecryptRequest:
		r.Options = opts.(*runtimev1.DecryptRequestOptions)
	}
}

// Reset will call the `Reset` method in the given runtime proto message.
func Reset[T runtimev1.DecryptRequest | runtimev1.EncryptRequest | runtimev1.DecryptResponse | runtimev1.EncryptResponse](msg *T) {
	if msg == nil {
		return
	}

	switch r := any(msg).(type) {
	case *runtimev1.EncryptRequest:
		r.Reset()
	case *runtimev1.DecryptRequest:
		r.Reset()
	case *runtimev1.EncryptResponse:
		r.Reset()
	case *runtimev1.DecryptResponse:
		r.Reset()
	}
}
