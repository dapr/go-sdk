package client

import (
	"context"
	"net"
	"os"
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

var (
	testClient Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	c, f := getTestClient(ctx)
	testClient = c
	r := m.Run()
	f()
	os.Exit(r)
}

func TestNewClientWithoutArgs(t *testing.T) {
	_, err := NewClientWithPort("")
	assert.NotNil(t, err)
	_, err = NewClientWithAddress("")
	assert.NotNil(t, err)
}

func getTestClient(ctx context.Context) (client Client, closer func()) {
	s := grpc.NewServer()
	pb.RegisterDaprServer(s, &testDaprServer{
		state: make(map[string][]byte),
	})

	l := bufconn.Listen(testBufSize)
	go func() {
		if err := s.Serve(l); err != nil {
			logger.Fatalf("test server exited with error: %v", err)
		}
	}()

	d := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	})

	c, err := grpc.DialContext(ctx, "", d, grpc.WithInsecure())
	if err != nil {
		logger.Fatalf("failed to dial test context: %v", err)
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
	d := make(map[string]string)
	d["test"] = "value"
	return &pb.GetSecretResponse{
		Data: d,
	}, nil
}
