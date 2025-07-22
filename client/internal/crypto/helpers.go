package crypto

import (
	"google.golang.org/protobuf/proto"

	commonv1pb "github.com/dapr/go-sdk/internal/proto/dapr/proto/common/v1"
	runtimev1 "github.com/dapr/go-sdk/internal/proto/dapr/proto/runtime/v1"
)

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
