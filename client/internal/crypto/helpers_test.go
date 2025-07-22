package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/dapr/go-sdk/client/internal/crypto"
	commonv1 "github.com/dapr/go-sdk/internal/proto/dapr/proto/common/v1"
	runtimev1 "github.com/dapr/go-sdk/internal/proto/dapr/proto/runtime/v1"
)

func TestPayloadMethods(t *testing.T) {
	testCases := map[string]struct {
		protoMsg  any
		inputData []byte
	}{
		"EncryptRequest": {
			protoMsg:  &runtimev1.EncryptRequest{},
			inputData: []byte("test data"),
		},
		"EncryptResponse": {
			protoMsg:  &runtimev1.EncryptResponse{},
			inputData: []byte("test data"),
		},
		"DecryptRequest": {
			protoMsg:  &runtimev1.DecryptRequest{},
			inputData: []byte("test data"),
		},
		"DecryptResponse": {
			protoMsg:  &runtimev1.DecryptResponse{},
			inputData: []byte("test data"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			inputPayload := &commonv1.StreamPayload{Data: tc.inputData}
			var outputPayload *commonv1.StreamPayload

			switch r := tc.protoMsg.(type) {
			case *runtimev1.EncryptRequest:
				crypto.SetPayload(r, inputPayload)
				outputPayload = crypto.GetPayload(r)
			case *runtimev1.EncryptResponse:
				crypto.SetPayload(r, inputPayload)
				outputPayload = crypto.GetPayload(r)
			case *runtimev1.DecryptRequest:
				crypto.SetPayload(r, inputPayload)
				outputPayload = crypto.GetPayload(r)
			case *runtimev1.DecryptResponse:
				crypto.SetPayload(r, inputPayload)
				outputPayload = crypto.GetPayload(r)
			default:
				require.Failf(t, "unsupported proto message type", "the type was %T", r)
			}

			assert.Equal(t, tc.inputData, outputPayload.GetData(), "payload should match the input")
		})
	}
}

func TestSetOptions(t *testing.T) {
	testCases := map[string]struct {
		protoMsg any
		options  proto.Message
	}{
		"EncryptRequest": {
			protoMsg: &runtimev1.EncryptRequest{},
			options: &runtimev1.EncryptRequestOptions{
				KeyName: "testing",
			},
		},
		"DecryptRequest": {
			protoMsg: &runtimev1.DecryptRequest{},
			options: &runtimev1.DecryptRequestOptions{
				KeyName: "testing",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var outputOptions proto.Message

			switch r := tc.protoMsg.(type) {
			case *runtimev1.EncryptRequest:
				crypto.SetOptions(r, tc.options)
				outputOptions = r.GetOptions()
			case *runtimev1.DecryptRequest:
				crypto.SetOptions(r, tc.options)
				outputOptions = r.GetOptions()
			default:
				require.Failf(t, "unsupported proto message type", "the type was %T", r)
			}

			assert.Equal(t, tc.options, outputOptions, "options should be persisted")
		})
	}
}

func TestReset(t *testing.T) {
	testCases := map[string]struct {
		protoMsg any
	}{
		"EncryptRequest": {
			protoMsg: &runtimev1.EncryptRequest{Payload: &commonv1.StreamPayload{Data: []byte("test data")}},
		},
		"EncryptResponse": {
			protoMsg: &runtimev1.EncryptResponse{Payload: &commonv1.StreamPayload{Data: []byte("test data")}},
		},
		"DecryptRequest": {
			protoMsg: &runtimev1.DecryptRequest{Payload: &commonv1.StreamPayload{Data: []byte("test data")}},
		},
		"DecryptResponse": {
			protoMsg: &runtimev1.DecryptResponse{Payload: &commonv1.StreamPayload{Data: []byte("test data")}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var payload *commonv1.StreamPayload

			switch r := tc.protoMsg.(type) {
			case *runtimev1.EncryptRequest:
				crypto.Reset(r)
				payload = crypto.GetPayload(r)
			case *runtimev1.EncryptResponse:
				crypto.Reset(r)
				payload = crypto.GetPayload(r)
			case *runtimev1.DecryptRequest:
				crypto.Reset(r)
				payload = crypto.GetPayload(r)
			case *runtimev1.DecryptResponse:
				crypto.Reset(r)
				payload = crypto.GetPayload(r)
			default:
				require.Failf(t, "unsupported proto message type", "the type was %T", r)
			}

			assert.Nil(t, payload)
		})
	}
}
