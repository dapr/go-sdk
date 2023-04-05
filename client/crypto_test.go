/*
Copyright 2021 The Dapr Authors
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
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"

	commonv1 "github.com/dapr/go-sdk/dapr/proto/common/v1"
	runtimev1pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

func (s *testDaprServer) EncryptAlpha1(stream runtimev1pb.Dapr_EncryptAlpha1Server) error {
	return s.performCryptoOperation(
		stream,
		&runtimev1pb.EncryptAlpha1Request{},
		&runtimev1pb.EncryptAlpha1Response{},
	)
}

func (s *testDaprServer) DecryptAlpha1(stream runtimev1pb.Dapr_DecryptAlpha1Server) error {
	return s.performCryptoOperation(
		stream,
		&runtimev1pb.DecryptAlpha1Request{},
		&runtimev1pb.DecryptAlpha1Response{},
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
				if payload.Seq != expectSeq {
					pw.CloseWithError(fmt.Errorf("invalid sequence number: %d (expected: %d)", payload.Seq, expectSeq))
					return
				}
				expectSeq++

				_, err = pw.Write(payload.Data)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
			}
		}
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
