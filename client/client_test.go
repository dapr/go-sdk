package client

import (
	"context"
	"net"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	commonv1pb "github.com/dapr/go-sdk/dapr/proto/common/v1"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
)

const (
	testBufSize = 1024 * 1024
)

func TestNewClientWithConnection(t *testing.T) {
	ctx := context.Background()
	client, closer := getTestClient(ctx, t)
	assert.NotNil(t, closer)
	defer closer()
	assert.NotNil(t, client)
}

func TestNewClientWithoutArgs(t *testing.T) {
	_, err := NewClientWithPort("")
	assert.NotNil(t, err)
	_, err = NewClientWithAddress("")
	assert.NotNil(t, err)
}

func getTestClient(ctx context.Context, t *testing.T) (client Client, closer func()) {
	l := bufconn.Listen(testBufSize)
	s := grpc.NewServer()

	server := &testDaprServer{
		state: make(map[string][]byte, 0),
	}

	pb.RegisterDaprServer(s, server)

	go func() {
		if err := s.Serve(l); err != nil {
			t.Fatalf("test server exited with error: %v", err)
		}
	}()

	d := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	})

	c, err := grpc.DialContext(ctx, "bufnet", d, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	closer = func() {
		l.Close()
		s.Stop()
	}

	client = NewClientWithConnection(c)
	return
}

type testDaprServer struct {
	state map[string][]byte
}

func (s *testDaprServer) InvokeService(ctx context.Context, req *pb.InvokeServiceRequest) (*commonv1pb.InvokeResponse, error) {
	return &commonv1pb.InvokeResponse{
		ContentType: req.Message.ContentType,
		Data:        req.GetMessage().Data,
	}, nil
}

func (s *testDaprServer) GetState(ctx context.Context, req *pb.GetStateRequest) (*pb.GetStateResponse, error) {
	return &pb.GetStateResponse{
		Data: s.state[req.Key],
		Etag: "v1",
	}, nil
}

func (s *testDaprServer) SaveState(ctx context.Context, req *pb.SaveStateRequest) (*empty.Empty, error) {
	for _, item := range req.States {
		s.state[item.Key] = item.Value
	}
	return &empty.Empty{}, nil
}

func (s *testDaprServer) DeleteState(ctx context.Context, req *pb.DeleteStateRequest) (*empty.Empty, error) {
	delete(s.state, req.Key)
	return &empty.Empty{}, nil
}

func (s *testDaprServer) PublishEvent(ctx context.Context, req *pb.PublishEventRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (s *testDaprServer) InvokeBinding(ctx context.Context, req *pb.InvokeBindingRequest) (*pb.InvokeBindingResponse, error) {
	return &pb.InvokeBindingResponse{
		Data:     req.Data,
		Metadata: req.Metadata,
	}, nil
}

func (s *testDaprServer) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	d := make(map[string]string, 0)
	d["test"] = "value"
	return &pb.GetSecretResponse{
		Data: d,
	}, nil
}
