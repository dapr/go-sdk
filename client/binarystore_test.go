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
	"bytes"
	"cmp"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	commonv1pb "github.com/dapr/dapr/pkg/proto/common/v1"
	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

const testBinaryStore = "binstore"

func TestSetBinaryFileAlpha1(t *testing.T) {
	ctx := t.Context()

	t.Run("invalid arguments fail without a call", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)

		err := c.SetBinaryFileAlpha1(ctx, "", "file", strings.NewReader("data"), false)
		require.EqualError(t, err, "storeName is empty")
		err = c.SetBinaryFileAlpha1(ctx, testBinaryStore, "", strings.NewReader("data"), false)
		require.EqualError(t, err, "fileName is empty")
		err = c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", nil, false)
		require.EqualError(t, err, "in is nil")
		assert.Zero(t, srv.calls.Load())
	})

	t.Run("sends the initial request and then the chunks in order", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)
		data := testBinaryData(3*binaryFileChunkSize + 7)

		err := c.SetBinaryFileAlpha1(ctx, testBinaryStore, "dir/file.bin", bytes.NewReader(data), false)
		require.NoError(t, err)

		stored, ok := srv.file(testBinaryStore, "dir/file.bin")
		require.True(t, ok)
		assert.Equal(t, data, stored)

		upload := srv.lastUpload()
		assert.Equal(t, testBinaryStore, upload.initial.GetComponentName())
		assert.Equal(t, "dir/file.bin", upload.initial.GetFileName())
		require.NotNil(t, upload.initial.Overwrite, "overwrite must always be sent")
		assert.False(t, upload.initial.GetOverwrite())
		assert.Equal(t, []int{binaryFileChunkSize, binaryFileChunkSize, binaryFileChunkSize, 7}, upload.chunkSizes)
	})

	t.Run("fills each chunk from a reader that returns little data at a time", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)
		data := testBinaryData(binaryFileChunkSize + 100)
		in := iotest.DataErrReader(iotest.OneByteReader(bytes.NewReader(data)))

		err := c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", in, false)
		require.NoError(t, err)

		stored, ok := srv.file(testBinaryStore, "file")
		require.True(t, ok)
		assert.Equal(t, data, stored)
		assert.Equal(t, []int{binaryFileChunkSize, 100}, srv.lastUpload().chunkSizes)
	})

	t.Run("empty file", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)

		err := c.SetBinaryFileAlpha1(ctx, testBinaryStore, "empty", strings.NewReader(""), false)
		require.NoError(t, err)

		stored, ok := srv.file(testBinaryStore, "empty")
		require.True(t, ok)
		assert.Empty(t, stored)
		assert.Empty(t, srv.lastUpload().chunkSizes)
	})

	t.Run("existing file without overwrite", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)
		err := c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", strings.NewReader("old"), false)
		require.NoError(t, err)

		// The server rejects the upload after the initial request, so with this much data the
		// client is usually still sending: the status must be returned, not io.EOF.
		err = c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", bytes.NewReader(testBinaryData(4<<20)), false)
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))

		stored, _ := srv.file(testBinaryStore, "file")
		assert.Equal(t, "old", string(stored))
	})

	t.Run("existing file with overwrite", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)
		err := c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", strings.NewReader("old"), false)
		require.NoError(t, err)

		err = c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", strings.NewReader("new"), true)
		require.NoError(t, err)

		stored, _ := srv.file(testBinaryStore, "file")
		assert.Equal(t, "new", string(stored))
		assert.True(t, srv.lastUpload().initial.GetOverwrite())
	})

	t.Run("a reader error cancels the upload", func(t *testing.T) {
		// io.ErrUnexpectedEOF is what a truncated source returns, such as a cut-short HTTP body.
		for _, errRead := range []error{errors.New("simulated read error"), io.ErrUnexpectedEOF} {
			t.Run(errRead.Error(), func(t *testing.T) {
				srv := &fakeBinaryStoreServer{}
				c := newTestBinaryStoreServer(t, srv)
				in := io.MultiReader(bytes.NewReader(testBinaryData(binaryFileChunkSize+10)), iotest.ErrReader(errRead))

				err := c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", in, false)
				require.ErrorIs(t, err, errRead)

				// The server must see a canceled stream, not the end of the upload, or it would store a partial file.
				assert.Equal(t, codes.Canceled, status.Code(receiveWithin(t, srv.setDone)))
				_, ok := srv.file(testBinaryStore, "file")
				assert.False(t, ok)
			})
		}
	})

	t.Run("canceling the context cancels the upload", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{received: make(chan int, 16)}
		c := newTestBinaryStoreServer(t, srv)
		uploadCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		release := make(chan struct{})
		releaseReader := sync.OnceFunc(func() { close(release) })
		// Unblocks the upload goroutine if the test fails early.
		t.Cleanup(releaseReader)
		in := io.MultiReader(
			bytes.NewReader(testBinaryData(binaryFileChunkSize)),
			readerFunc(func([]byte) (int, error) {
				<-release
				return 0, io.EOF
			}),
		)

		errCh := make(chan error, 1)
		go func() {
			errCh <- c.SetBinaryFileAlpha1(uploadCtx, testBinaryStore, "file", in, false)
		}()

		receiveWithin(t, srv.received)
		cancel()
		assert.Equal(t, codes.Canceled, status.Code(receiveWithin(t, srv.setDone)))
		releaseReader()

		err := receiveWithin(t, errCh)
		assert.Equal(t, codes.Canceled, status.Code(err))
		_, ok := srv.file(testBinaryStore, "file")
		assert.False(t, ok)
	})
}

func TestGetBinaryFileAlpha1(t *testing.T) {
	ctx := t.Context()

	t.Run("invalid arguments fail without a call", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, "", "file")
		require.EqualError(t, err, "storeName is empty")
		assert.Nil(t, r)
		r, err = c.GetBinaryFileAlpha1(ctx, testBinaryStore, "")
		require.EqualError(t, err, "fileName is empty")
		assert.Nil(t, r)
		assert.Zero(t, srv.calls.Load())
	})

	t.Run("reads every chunk", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{getChunkSize: 1000}
		c := newTestBinaryStoreServer(t, srv)
		data := testBinaryData(3*binaryFileChunkSize + 7)
		require.NoError(t, c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", bytes.NewReader(data), false))

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		require.NoError(t, err)
		require.NoError(t, iotest.TestReader(r, data))
		require.NoError(t, r.Close())
	})

	t.Run("empty file", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)
		require.NoError(t, c.SetBinaryFileAlpha1(ctx, testBinaryStore, "empty", strings.NewReader(""), false))

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "empty")
		require.NoError(t, err)
		require.NoError(t, iotest.TestReader(r, nil))
		require.NoError(t, r.Close())
	})

	t.Run("missing file fails before returning a reader", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "missing")
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Nil(t, r)
	})

	t.Run("an error after the first chunk keeps its code", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{
			getHandler: func(_ *runtimev1pb.GetBinaryFileRequestAlpha1, stream runtimev1pb.Dapr_GetBinaryFileAlpha1Server) error {
				if err := sendBinaryChunk(stream, 0, "abc"); err != nil {
					return err
				}
				return status.Error(codes.Internal, "simulated failure")
			},
		}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		require.NoError(t, err)
		data, err := io.ReadAll(r)
		assert.Equal(t, "abc", string(data))
		require.ErrorContains(t, err, "simulated failure")
		assert.Equal(t, codes.Internal, status.Code(err))
		require.NoError(t, r.Close())
	})

	t.Run("a chunk out of order is an error", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{
			getHandler: func(_ *runtimev1pb.GetBinaryFileRequestAlpha1, stream runtimev1pb.Dapr_GetBinaryFileAlpha1Server) error {
				if err := sendBinaryChunk(stream, 0, "a"); err != nil {
					return err
				}
				return sendBinaryChunk(stream, 2, "b")
			},
		}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		require.NoError(t, err)
		data, err := io.ReadAll(r)
		assert.Equal(t, "a", string(data))
		require.EqualError(t, err, "invalid sequence number in chunk: 2 (expected: 1)")
		_, err = r.Read(make([]byte, 1))
		require.EqualError(t, err, "invalid sequence number in chunk: 2 (expected: 1)")
		require.NoError(t, r.Close())
	})

	t.Run("a first chunk out of order fails before returning a reader", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{
			getHandler: func(_ *runtimev1pb.GetBinaryFileRequestAlpha1, stream runtimev1pb.Dapr_GetBinaryFileAlpha1Server) error {
				return sendBinaryChunk(stream, 1, "a")
			},
		}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		require.EqualError(t, err, "invalid sequence number in chunk: 1 (expected: 0)")
		assert.Nil(t, r)
	})

	t.Run("closing the reader early cancels the download", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{getHandler: sendBinaryChunkAndWait("first chunk")}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		require.NoError(t, err)
		buf := make([]byte, 5)
		n, err := r.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, "first", string(buf[:n]))

		require.NoError(t, r.Close())
		require.ErrorIs(t, receiveWithin(t, srv.getDone), context.Canceled)
		_, err = r.Read(buf)
		require.ErrorIs(t, err, os.ErrClosed)
		require.NoError(t, r.Close())
	})

	t.Run("closing the reader from another goroutine stops its read", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{getHandler: sendBinaryChunkAndWait("first chunk")}
		c := newTestBinaryStoreServer(t, srv)

		r, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		require.NoError(t, err)
		buf := make([]byte, len("first chunk"))
		_, err = io.ReadFull(r, buf)
		require.NoError(t, err)
		// The server is waiting, so this read blocks until Close.
		blockedRead := make(chan error, 1)
		go func() {
			_, readErr := r.Read(buf)
			blockedRead <- readErr
		}()

		require.NoError(t, r.Close())
		require.ErrorIs(t, receiveWithin(t, blockedRead), os.ErrClosed)
		require.ErrorIs(t, receiveWithin(t, srv.getDone), context.Canceled)
	})

	t.Run("canceling the context cancels the download", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{getHandler: sendBinaryChunkAndWait("first chunk")}
		c := newTestBinaryStoreServer(t, srv)
		downloadCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		r, err := c.GetBinaryFileAlpha1(downloadCtx, testBinaryStore, "file")
		require.NoError(t, err)
		buf := make([]byte, len("first chunk"))
		_, err = io.ReadFull(r, buf)
		require.NoError(t, err)
		// The server is waiting, so an empty read that received the next chunk would block.
		emptyRead := make(chan error, 1)
		go func() {
			_, readErr := r.Read(nil)
			emptyRead <- readErr
		}()
		require.NoError(t, receiveWithin(t, emptyRead))

		cancel()
		_, err = r.Read(buf)
		assert.Equal(t, codes.Canceled, status.Code(err))
		require.ErrorIs(t, receiveWithin(t, srv.getDone), context.Canceled)
		require.NoError(t, r.Close())
	})
}

func TestDeleteBinaryFileAlpha1(t *testing.T) {
	ctx := t.Context()

	t.Run("invalid arguments fail without a call", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)

		require.EqualError(t, c.DeleteBinaryFileAlpha1(ctx, "", "file"), "storeName is empty")
		require.EqualError(t, c.DeleteBinaryFileAlpha1(ctx, testBinaryStore, ""), "fileName is empty")
		assert.Zero(t, srv.calls.Load())
	})

	t.Run("deletes the file", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)
		require.NoError(t, c.SetBinaryFileAlpha1(ctx, testBinaryStore, "file", strings.NewReader("data"), false))

		require.NoError(t, c.DeleteBinaryFileAlpha1(ctx, testBinaryStore, "file"))

		_, ok := srv.file(testBinaryStore, "file")
		assert.False(t, ok)
		_, err := c.GetBinaryFileAlpha1(ctx, testBinaryStore, "file")
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("missing file", func(t *testing.T) {
		srv := &fakeBinaryStoreServer{}
		c := newTestBinaryStoreServer(t, srv)

		err := c.DeleteBinaryFileAlpha1(ctx, testBinaryStore, "missing")
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

// newTestBinaryStoreServer starts srv on a dedicated in-memory connection and returns a client
// connected to it. Set srv's configuration fields before calling it.
func newTestBinaryStoreServer(t *testing.T, srv *fakeBinaryStoreServer) Client {
	t.Helper()

	srv.files = make(map[string][]byte)
	srv.setDone = make(chan error, 16)
	srv.getDone = make(chan error, 16)

	s := grpc.NewServer()
	runtimev1pb.RegisterDaprServer(s, srv)

	l := bufconn.Listen(testBufSize)
	go func() {
		_ = s.Serve(l)
	}()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return l.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		s.Stop()
		_ = l.Close()
	})

	return NewClientWithConnection(conn)
}

// fakeBinaryStoreServer is an in-memory binary store. Like the runtime, it requires the initial
// request first and only once, and seq from 0, and it rejects an existing file before reading
// the data. It doesn't validate file names or time out waiting for the first message.
type fakeBinaryStoreServer struct {
	runtimev1pb.UnimplementedDaprServer

	// Size of the chunks sent by GetBinaryFileAlpha1; defaults to binaryFileChunkSize.
	getChunkSize int
	// If set, replaces the default GetBinaryFileAlpha1 behavior.
	getHandler func(*runtimev1pb.GetBinaryFileRequestAlpha1, runtimev1pb.Dapr_GetBinaryFileAlpha1Server) error
	// If set, receives the size of each chunk SetBinaryFileAlpha1 receives.
	received chan int

	calls   atomic.Int32
	setDone chan error // Receives the result of each SetBinaryFileAlpha1 call.
	getDone chan error // Receives the result of each GetBinaryFileAlpha1 call.

	lock    sync.Mutex
	files   map[string][]byte
	uploads []fakeBinaryUpload
}

type fakeBinaryUpload struct {
	initial    *runtimev1pb.SetBinaryFileRequestInitialAlpha1
	chunkSizes []int
}

func (s *fakeBinaryStoreServer) SetBinaryFileAlpha1(stream runtimev1pb.Dapr_SetBinaryFileAlpha1Server) (err error) {
	s.calls.Add(1)
	var upload fakeBinaryUpload
	defer func() {
		s.lock.Lock()
		s.uploads = append(s.uploads, upload)
		s.lock.Unlock()
		notify(s.setDone, err)
	}()

	req, err := stream.Recv()
	if err != nil {
		return err
	}
	upload.initial = req.GetInitialRequest()
	if upload.initial == nil {
		return status.Error(codes.InvalidArgument, "first message does not contain the initial request")
	}
	key := upload.initial.GetComponentName() + "/" + upload.initial.GetFileName()
	if _, exists := s.file(upload.initial.GetComponentName(), upload.initial.GetFileName()); exists && !upload.initial.GetOverwrite() {
		return status.Errorf(codes.AlreadyExists, "file %s already exists", upload.initial.GetFileName())
	}

	var data bytes.Buffer
	for seq := uint64(0); ; seq++ {
		req, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if req.GetInitialRequest() != nil {
			return status.Error(codes.InvalidArgument, "initial request found in non-leading message")
		}
		payload := req.GetPayload()
		if payload.GetSeq() != seq {
			return status.Errorf(codes.InvalidArgument, "invalid sequence number received: %d (expected: %d)", payload.GetSeq(), seq)
		}
		upload.chunkSizes = append(upload.chunkSizes, len(payload.GetData()))
		data.Write(payload.GetData())
		notify(s.received, len(payload.GetData()))
	}

	s.lock.Lock()
	s.files[key] = data.Bytes()
	s.lock.Unlock()
	return stream.SendAndClose(&runtimev1pb.SetBinaryFileResponseAlpha1{})
}

func (s *fakeBinaryStoreServer) GetBinaryFileAlpha1(req *runtimev1pb.GetBinaryFileRequestAlpha1, stream runtimev1pb.Dapr_GetBinaryFileAlpha1Server) (err error) {
	s.calls.Add(1)
	defer func() {
		notify(s.getDone, err)
	}()

	if s.getHandler != nil {
		return s.getHandler(req, stream)
	}

	data, ok := s.file(req.GetComponentName(), req.GetFileName())
	if !ok {
		return status.Errorf(codes.NotFound, "file %s not found", req.GetFileName())
	}
	chunkSize := cmp.Or(s.getChunkSize, binaryFileChunkSize)
	for seq := uint64(0); len(data) > 0; seq++ {
		n := min(chunkSize, len(data))
		if err = sendBinaryChunk(stream, seq, string(data[:n])); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

func (s *fakeBinaryStoreServer) DeleteBinaryFileAlpha1(_ context.Context, req *runtimev1pb.DeleteBinaryFileRequestAlpha1) (*runtimev1pb.DeleteBinaryFileResponseAlpha1, error) {
	s.calls.Add(1)
	key := req.GetComponentName() + "/" + req.GetFileName()

	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.files[key]; !ok {
		return nil, status.Errorf(codes.NotFound, "file %s not found", req.GetFileName())
	}
	delete(s.files, key)
	return &runtimev1pb.DeleteBinaryFileResponseAlpha1{}, nil
}

func (s *fakeBinaryStoreServer) file(storeName, fileName string) ([]byte, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	data, ok := s.files[storeName+"/"+fileName]
	return data, ok
}

func (s *fakeBinaryStoreServer) lastUpload() fakeBinaryUpload {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.uploads[len(s.uploads)-1]
}

func sendBinaryChunk(stream runtimev1pb.Dapr_GetBinaryFileAlpha1Server, seq uint64, data string) error {
	return stream.Send(&runtimev1pb.GetBinaryFileResponseAlpha1{
		Payload: &commonv1pb.StreamPayload{
			Data: []byte(data),
			Seq:  seq,
		},
	})
}

// sendBinaryChunkAndWait returns a GetBinaryFileAlpha1 handler that sends one chunk and then
// waits until the client cancels the stream.
func sendBinaryChunkAndWait(data string) func(*runtimev1pb.GetBinaryFileRequestAlpha1, runtimev1pb.Dapr_GetBinaryFileAlpha1Server) error {
	return func(_ *runtimev1pb.GetBinaryFileRequestAlpha1, stream runtimev1pb.Dapr_GetBinaryFileAlpha1Server) error {
		if err := sendBinaryChunk(stream, 0, data); err != nil {
			return err
		}
		<-stream.Context().Done()
		return stream.Context().Err()
	}
}

// notify sends v on ch without blocking; ch may be nil.
func notify[T any](ch chan T, v T) {
	select {
	case ch <- v:
	default:
	}
}

// receiveWithin returns the next value from ch, failing the test if none arrives in time.
func receiveWithin[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(10 * time.Second):
		require.FailNow(t, "timed out waiting for the test server")
		var zero T
		return zero
	}
}

// testBinaryData returns n bytes whose pattern doesn't line up with the chunk size, so data in
// the wrong order is detected.
func testBinaryData(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i % 251)
	}
	return data
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) {
	return f(p)
}
