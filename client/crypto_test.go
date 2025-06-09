/*
Copyright 2023 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	commonv1 "github.com/dapr/dapr/pkg/proto/common/v1"
	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

func TestEncrypt(t *testing.T) {
	ctx := t.Context()

	t.Run("missing ComponentName", func(t *testing.T) {
		out, err := testClient.Encrypt(ctx,
			strings.NewReader("hello world"),
			EncryptOptions{
				// ComponentName: "mycomponent",
				KeyName:          "key",
				KeyWrapAlgorithm: "algorithm",
			},
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "ComponentName")
		require.Nil(t, out)
	})

	t.Run("missing Key", func(t *testing.T) {
		out, err := testClient.Encrypt(ctx,
			strings.NewReader("hello world"),
			EncryptOptions{
				ComponentName: "mycomponent",
				// KeyName:       "key",
				KeyWrapAlgorithm: "algorithm",
			},
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "Key")
		require.Nil(t, out)
	})

	t.Run("missing Algorithm", func(t *testing.T) {
		out, err := testClient.Encrypt(ctx,
			strings.NewReader("hello world"),
			EncryptOptions{
				ComponentName: "mycomponent",
				KeyName:       "key",
				// Algorithm: "algorithm",
			},
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "Algorithm")
		require.Nil(t, out)
	})

	t.Run("receiving back data sent", func(t *testing.T) {
		// The test server doesn't actually encrypt data
		out, err := testClient.Encrypt(ctx,
			strings.NewReader("hello world"),
			EncryptOptions{
				ComponentName:    "mycomponent",
				KeyName:          "key",
				KeyWrapAlgorithm: "algorithm",
			},
		)
		require.NoError(t, err)
		require.NotNil(t, out)

		read, err := io.ReadAll(out)
		require.NoError(t, err)
		require.Equal(t, "hello world", string(read))
	})

	t.Run("error in input stream", func(t *testing.T) {
		out, err := testClient.Encrypt(ctx,
			&failingReader{
				data: strings.NewReader("hello world"),
			},
			EncryptOptions{
				ComponentName:    "mycomponent",
				KeyName:          "key",
				KeyWrapAlgorithm: "algorithm",
			},
		)
		require.NoError(t, err)
		require.NotNil(t, out)

		_, err = io.ReadAll(out)
		require.Error(t, err)
		require.ErrorContains(t, err, "simulated")
	})

	t.Run("context canceled", func(t *testing.T) {
		failingCtx, failingCancel := context.WithTimeout(ctx, time.Second)
		defer failingCancel()

		out, err := testClient.Encrypt(failingCtx,
			&slowReader{
				// Should take a lot longer than 1s
				//nolint:dupword
				data:  strings.NewReader("soft kitty, warm kitty, little ball of fur, happy kitty, sleepy kitty, purr purr purr"),
				delay: time.Second,
			},
			EncryptOptions{
				ComponentName:    "mycomponent",
				KeyName:          "key",
				KeyWrapAlgorithm: "algorithm",
			},
		)
		require.NoError(t, err)
		require.NotNil(t, out)

		_, err = io.ReadAll(out)
		require.Error(t, err)
		require.ErrorContains(t, err, "context deadline exceeded")
	})
}

func TestDecrypt(t *testing.T) {
	ctx := t.Context()

	t.Run("missing ComponentName", func(t *testing.T) {
		out, err := testClient.Decrypt(ctx,
			strings.NewReader("hello world"),
			DecryptOptions{
				// ComponentName: "mycomponent",
			},
		)
		require.Error(t, err)
		require.ErrorContains(t, err, "ComponentName")
		require.Nil(t, out)
	})

	t.Run("receiving back data sent", func(t *testing.T) {
		// The test server doesn't actually decrypt data
		out, err := testClient.Decrypt(ctx,
			strings.NewReader("hello world"),
			DecryptOptions{
				ComponentName: "mycomponent",
			},
		)
		require.NoError(t, err)
		require.NotNil(t, out)

		read, err := io.ReadAll(out)
		require.NoError(t, err)
		require.Equal(t, "hello world", string(read))
	})

	t.Run("error in input stream", func(t *testing.T) {
		out, err := testClient.Decrypt(ctx,
			&failingReader{
				data: strings.NewReader("hello world"),
			},
			DecryptOptions{
				ComponentName: "mycomponent",
			},
		)
		require.NoError(t, err)
		require.NotNil(t, out)

		_, err = io.ReadAll(out)
		require.Error(t, err)
		require.ErrorContains(t, err, "simulated")
	})
}

/* --- Server methods --- */

func (s *testDaprServer) EncryptAlpha1(stream runtimev1pb.Dapr_EncryptAlpha1Server) error {
	return s.performCryptoOperation(
		stream,
		&runtimev1pb.EncryptRequest{},
		&runtimev1pb.EncryptResponse{},
	)
}

func (s *testDaprServer) DecryptAlpha1(stream runtimev1pb.Dapr_DecryptAlpha1Server) error {
	return s.performCryptoOperation(
		stream,
		&runtimev1pb.DecryptRequest{},
		&runtimev1pb.DecryptResponse{},
	)
}

func (s *testDaprServer) performCryptoOperation(stream grpc.ServerStream, reqProto runtimev1pb.CryptoRequests, resProto runtimev1pb.CryptoResponses) error {
	// This doesn't really encrypt or decrypt the data and just sends back whatever it receives
	pr, pw := io.Pipe()

	go func() {
		var (
			done      bool
			err       error
			expectSeq uint64
		)
		first := true
		for !done && stream.Context().Err() == nil {
			reqProto.Reset()
			err = stream.RecvMsg(reqProto)
			if errors.Is(err, io.EOF) {
				done = true
			} else if err != nil {
				pw.CloseWithError(err)
				return
			}

			if first && !reqProto.HasOptions() {
				pw.CloseWithError(errors.New("first message must have options"))
				return
			} else if !first && reqProto.HasOptions() {
				pw.CloseWithError(errors.New("messages after first must not have options"))
				return
			}
			first = false

			payload := reqProto.GetPayload()
			if payload != nil {
				if payload.GetSeq() != expectSeq {
					pw.CloseWithError(fmt.Errorf("invalid sequence number: %d (expected: %d)", payload.GetSeq(), expectSeq))
					return
				}
				expectSeq++

				_, err = pw.Write(payload.GetData())
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			}
		}

		pw.Close()
	}()

	var (
		done bool
		n    int
		err  error
		seq  uint64
	)
	buf := make([]byte, 2<<10)
	for !done && stream.Context().Err() == nil {
		resProto.Reset()

		n, err = pr.Read(buf)
		if errors.Is(err, io.EOF) {
			done = true
		} else if err != nil {
			return err
		}

		if n > 0 {
			resProto.SetPayload(&commonv1.StreamPayload{
				Seq:  seq,
				Data: buf[:n],
			})
			seq++

			err = stream.SendMsg(resProto)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
