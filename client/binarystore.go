/*
Copyright 2026 The Dapr Authors
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
	"os"
	"sync/atomic"

	commonv1pb "github.com/dapr/dapr/pkg/proto/common/v1"
	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

// Size of the chunks sent by SetBinaryFileAlpha1: 32KB, the same size the runtime uses for downloads.
const binaryFileChunkSize = 32 << 10

// SetBinaryFileAlpha1 stores the data read from in as fileName in the binary store storeName.
// If overwrite is false and the file already exists, the returned error has code codes.AlreadyExists.
// If reading from in fails, the upload is canceled instead of completed, and the read error is returned.
// Canceling ctx doesn't interrupt a blocked in.Read call.
// This is an alpha API and may change.
func (c *GRPCClient) SetBinaryFileAlpha1(ctx context.Context, storeName, fileName string, in io.Reader, overwrite bool) error {
	if storeName == "" {
		return errors.New("storeName is empty")
	}
	if fileName == "" {
		return errors.New("fileName is empty")
	}
	if in == nil {
		return errors.New("in is nil")
	}

	// Returning without CloseAndRecv cancels the stream, so the runtime doesn't store a partial file.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := c.protoClient.SetBinaryFileAlpha1(ctx)
	if err != nil {
		return fmt.Errorf("error setting binary file: %w", err)
	}

	err = stream.Send(&runtimev1pb.SetBinaryFileRequestAlpha1{
		SetBinaryFileRequestType: &runtimev1pb.SetBinaryFileRequestAlpha1_InitialRequest{
			InitialRequest: &runtimev1pb.SetBinaryFileRequestInitialAlpha1{
				ComponentName: storeName,
				FileName:      fileName,
				Overwrite:     &overwrite,
			},
		},
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error sending binary file request: %w", err)
	}

	// io.EOF from Send means the runtime has ended the stream (for example because the file
	// exists); the error is then returned by CloseAndRecv.
	var seq uint64
	for err == nil {
		// gRPC may still hold a message after Send returns, so each chunk gets its own buffer.
		chunk := make([]byte, binaryFileChunkSize)
		n, readErr := readBinaryFileChunk(in, chunk)
		if readErr != nil && readErr != io.EOF {
			return fmt.Errorf("error reading binary file data: %w", readErr)
		}
		if n > 0 {
			err = stream.Send(&runtimev1pb.SetBinaryFileRequestAlpha1{
				SetBinaryFileRequestType: &runtimev1pb.SetBinaryFileRequestAlpha1_Payload{
					Payload: &commonv1pb.StreamPayload{
						Data: chunk[:n],
						Seq:  seq,
					},
				},
			})
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("error sending binary file chunk: %w", err)
			}
			seq++
		}
		if readErr != nil {
			break
		}
	}

	if _, err = stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("error setting binary file: %w", err)
	}

	return nil
}

// readBinaryFileChunk fills chunk from in. Unlike io.ReadFull, it returns io.ErrUnexpectedEOF
// only when in returns it, so a truncated source isn't stored as a complete file.
func readBinaryFileChunk(in io.Reader, chunk []byte) (int, error) {
	var n int
	for n < len(chunk) {
		m, err := in.Read(chunk[n:])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// GetBinaryFileAlpha1 returns a reader that streams fileName from the binary store storeName.
// Errors that happen before the first chunk, such as codes.NotFound for a missing file, are
// returned by this method. ctx applies to the whole download, so canceling it makes the next Read
// fail. The caller must Close the reader. Close cancels the download if it hasn't finished, and it
// can be called while another goroutine is blocked in Read.
// This is an alpha API and may change.
func (c *GRPCClient) GetBinaryFileAlpha1(ctx context.Context, storeName, fileName string) (io.ReadCloser, error) {
	if storeName == "" {
		return nil, errors.New("storeName is empty")
	}
	if fileName == "" {
		return nil, errors.New("fileName is empty")
	}

	ctx, cancel := context.WithCancel(ctx)
	stream, err := c.protoClient.GetBinaryFileAlpha1(ctx, &runtimev1pb.GetBinaryFileRequestAlpha1{
		ComponentName: storeName,
		FileName:      fileName,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error getting binary file: %w", err)
	}

	// The runtime opens the file before it sends anything, so receiving the first chunk now
	// returns errors from this method instead of from the first Read.
	r := &binaryFileReader{stream: stream, cancel: cancel}
	r.err = r.next()
	if r.err != nil && !errors.Is(r.err, io.EOF) {
		return nil, r.err
	}

	return r, nil
}

// DeleteBinaryFileAlpha1 deletes fileName from the binary store storeName.
// If the file doesn't exist, the returned error has code codes.NotFound.
// This is an alpha API and may change.
func (c *GRPCClient) DeleteBinaryFileAlpha1(ctx context.Context, storeName, fileName string) error {
	if storeName == "" {
		return errors.New("storeName is empty")
	}
	if fileName == "" {
		return errors.New("fileName is empty")
	}

	_, err := c.protoClient.DeleteBinaryFileAlpha1(ctx, &runtimev1pb.DeleteBinaryFileRequestAlpha1{
		ComponentName: storeName,
		FileName:      fileName,
	})
	if err != nil {
		return fmt.Errorf("error deleting binary file: %w", err)
	}

	return nil
}

// binaryFileReader reads the chunks of a GetBinaryFileAlpha1 stream. Read isn't safe for
// concurrent use, but Close can be called at any time, so it only touches closed and cancel.
type binaryFileReader struct {
	stream  runtimev1pb.Dapr_GetBinaryFileAlpha1Client
	cancel  context.CancelFunc
	closed  atomic.Bool
	buf     []byte // Unread data of the current chunk.
	nextSeq uint64
	err     error // Returned once buf is empty: io.EOF or a stream error.
}

func (r *binaryFileReader) Read(p []byte) (int, error) {
	for {
		if r.closed.Load() {
			return 0, os.ErrClosed
		}
		if len(r.buf) > 0 {
			break
		}
		if r.err != nil {
			return 0, r.err
		}
		if len(p) == 0 {
			return 0, nil
		}
		r.err = r.next()
	}

	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// Close cancels the download if it hasn't finished. It always returns nil.
func (r *binaryFileReader) Close() error {
	r.closed.Store(true)
	r.cancel()
	return nil
}

// next receives the next chunk into buf. It returns io.EOF at the end of the file.
// The stream is released when it returns an error.
func (r *binaryFileReader) next() error {
	res, err := r.stream.Recv()
	if errors.Is(err, io.EOF) {
		r.cancel()
		return io.EOF
	}
	if err != nil {
		r.cancel()
		return fmt.Errorf("error getting binary file: %w", err)
	}

	payload := res.GetPayload()
	if payload.GetSeq() != r.nextSeq {
		r.cancel()
		return fmt.Errorf("invalid sequence number in chunk: %d (expected: %d)", payload.GetSeq(), r.nextSeq)
	}
	r.nextSeq++
	r.buf = payload.GetData()

	return nil
}
