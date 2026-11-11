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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	commonv1pb "github.com/dapr/dapr/pkg/proto/common/v1"
	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

const (
	testBufSize           = 1024 * 1024
	testSocket            = "/tmp/dapr.socket"
	testWorkflowFailureID = "test_failure_id"
)

var testClient Client

func TestMain(m *testing.M) {
	ctx := context.Background()
	c, f := getTestClient(ctx)
	testClient = c
	r := m.Run()
	f()

	if r != 0 {
		os.Exit(r)
	}

	c, f = getTestClientWithSocket(ctx)
	testClient = c
	r = m.Run()
	f()
	os.Exit(r)
}

func TestNewClient(t *testing.T) {
	t.Run("return error when unable to reach server", func(t *testing.T) {
		_, err := NewClientWithPort("1")
		require.Error(t, err)
	})

	t.Run("no arg for with port", func(t *testing.T) {
		_, err := NewClientWithPort("")
		require.Error(t, err)
	})

	t.Run("no arg for with address", func(t *testing.T) {
		_, err := NewClientWithAddress("")
		require.Error(t, err)
	})

	t.Run("no arg with socket", func(t *testing.T) {
		_, err := NewClientWithSocket("")
		require.Error(t, err)
	})

	t.Run("new client closed with token", func(t *testing.T) {
		t.Setenv(apiTokenEnvVarName, "test")
		c, err := NewClientWithSocket(testSocket)
		require.NoError(t, err)
		defer c.Close()
		c.WithAuthToken("")
	})

	t.Run("new client closed with empty token", func(t *testing.T) {
		testClient.WithAuthToken("")
	})

	t.Run("new client with trace ID", func(t *testing.T) {
		_ = testClient.WithTraceID(t.Context(), "test")
	})

	t.Run("new socket client closed with token", func(t *testing.T) {
		t.Setenv(apiTokenEnvVarName, "test")
		c, err := NewClientWithSocket(testSocket)
		require.NoError(t, err)
		defer c.Close()
		c.WithAuthToken("")
	})

	t.Run("new socket client closed with empty token", func(t *testing.T) {
		c, err := NewClientWithSocket(testSocket)
		require.NoError(t, err)
		defer c.Close()
		c.WithAuthToken("")
	})

	t.Run("new socket client with trace ID", func(t *testing.T) {
		c, err := NewClientWithSocket(testSocket)
		require.NoError(t, err)
		defer c.Close()
		ctx := c.WithTraceID(t.Context(), "")
		_ = c.WithTraceID(ctx, "test")
	})

	t.Run("new client with extra dial options", func(t *testing.T) {
		_, err := os.Stat(testSocket)
		if err != nil {
			return
		}

		c, err := NewClientWithSocket(testSocket, grpc.WithUserAgent("test"))
		require.NoError(t, err)
		defer c.Close()

		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()

		addr := "unix:" + testSocket
		c, err = NewClientWithAddressContext(ctx, addr, grpc.WithUserAgent("test"))
		require.NoError(t, err)
		defer c.Close()

		t.Setenv(daprGRPCEndpointEnvVarName, addr)
		c, err = NewClient(grpc.WithUserAgent("test"))
		require.NoError(t, err)
		defer c.Close()
	})
}

func TestShutdown(t *testing.T) {
	ctx := t.Context()

	t.Run("shutdown", func(t *testing.T) {
		err := testClient.Shutdown(ctx)
		require.NoError(t, err)
	})
}

func getTestClient(ctx context.Context) (client Client, closer func()) {
	s := grpc.NewServer()
	pb.RegisterDaprServer(s, &testDaprServer{
		state:                       make(map[string][]byte),
		configurationSubscriptionID: map[string]chan struct{}{},
	})

	l := bufconn.Listen(testBufSize)
	go func() {
		if err := s.Serve(l); err != nil && err.Error() != "closed" {
			logger.Fatalf("test server exited with error: %v", err)
		}
	}()

	d := grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	})
	//nolint:staticcheck
	c, err := grpc.DialContext(ctx, "", d, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func getTestClientWithSocket(ctx context.Context) (client Client, closer func()) {
	s := grpc.NewServer()
	pb.RegisterDaprServer(s, &testDaprServer{
		state:                       make(map[string][]byte),
		configurationSubscriptionID: map[string]chan struct{}{},
	})

	var lc net.ListenConfig
	l, err := lc.Listen(ctx, "unix", testSocket)
	if err != nil {
		logger.Fatalf("socket test server created with error: %v", err)
	}

	go func() {
		if err = s.Serve(l); err != nil && err.Error() != "accept unix /tmp/dapr.socket: use of closed network connection" {
			logger.Fatalf("socket test server exited with error: %v", err)
		}
	}()

	closer = func() {
		l.Close()
		s.Stop()
		os.Remove(testSocket)
	}

	if client, err = NewClientWithSocket(testSocket); err != nil {
		logger.Fatalf("socket test client created with error: %v", err)
	}

	return
}

func Test_getClientTimeoutSeconds(t *testing.T) {
	t.Run("empty env var", func(t *testing.T) {
		t.Setenv(clientTimeoutSecondsEnvVarName, "")
		got, err := getClientTimeoutSeconds()
		require.NoError(t, err)
		assert.Equal(t, clientDefaultTimeoutSeconds, got)
	})

	t.Run("invalid env var", func(t *testing.T) {
		t.Setenv(clientTimeoutSecondsEnvVarName, "invalid")
		_, err := getClientTimeoutSeconds()
		require.Error(t, err)
	})

	t.Run("normal env var", func(t *testing.T) {
		t.Setenv(clientTimeoutSecondsEnvVarName, "7")
		got, err := getClientTimeoutSeconds()
		require.NoError(t, err)
		assert.Equal(t, 7, got)
	})

	t.Run("zero env var", func(t *testing.T) {
		t.Setenv(clientTimeoutSecondsEnvVarName, "0")
		_, err := getClientTimeoutSeconds()
		require.Error(t, err)
	})

	t.Run("negative env var", func(t *testing.T) {
		t.Setenv(clientTimeoutSecondsEnvVarName, "-3")
		_, err := getClientTimeoutSeconds()
		require.Error(t, err)
	})
}

type testDaprServer struct {
	pb.UnimplementedDaprServer
	state                             map[string][]byte
	configurationSubscriptionIDMapLoc sync.Mutex
	configurationSubscriptionID       map[string]chan struct{}
}

func (s *testDaprServer) TryLockAlpha1(ctx context.Context, req *pb.TryLockRequest) (*pb.TryLockResponse, error) {
	return &pb.TryLockResponse{
		Success: true,
	}, nil
}

func (s *testDaprServer) UnlockAlpha1(ctx context.Context, req *pb.UnlockRequest) (*pb.UnlockResponse, error) {
	return &pb.UnlockResponse{
		Status: pb.UnlockResponse_SUCCESS,
	}, nil
}

func (s *testDaprServer) InvokeService(ctx context.Context, req *pb.InvokeServiceRequest) (*commonv1pb.InvokeResponse, error) {
	if req.GetMessage() == nil {
		return &commonv1pb.InvokeResponse{
			ContentType: "text/plain",
			Data: &anypb.Any{
				Value: []byte("pong"),
			},
		}, nil
	}
	return &commonv1pb.InvokeResponse{
		ContentType: req.GetMessage().GetContentType(),
		Data:        req.GetMessage().GetData(),
	}, nil
}

func (s *testDaprServer) GetState(ctx context.Context, req *pb.GetStateRequest) (*pb.GetStateResponse, error) {
	return &pb.GetStateResponse{
		Data: s.state[req.GetKey()],
		Etag: "1",
	}, nil
}

func (s *testDaprServer) GetBulkState(ctx context.Context, in *pb.GetBulkStateRequest) (*pb.GetBulkStateResponse, error) {
	items := make([]*pb.BulkStateItem, 0)
	for _, k := range in.GetKeys() {
		if v, found := s.state[k]; found {
			item := &pb.BulkStateItem{
				Key:  k,
				Etag: "1",
				Data: v,
			}
			items = append(items, item)
		}
	}
	return &pb.GetBulkStateResponse{
		Items: items,
	}, nil
}

func (s *testDaprServer) SaveState(ctx context.Context, req *pb.SaveStateRequest) (*emptypb.Empty, error) {
	for _, item := range req.GetStates() {
		s.state[item.GetKey()] = item.GetValue()
	}
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) QueryStateAlpha1(ctx context.Context, req *pb.QueryStateRequest) (*pb.QueryStateResponse, error) {
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(req.GetQuery()), &v); err != nil {
		return nil, err
	}

	ret := &pb.QueryStateResponse{
		Results: make([]*pb.QueryStateItem, 0, len(s.state)),
	}
	for key, value := range s.state {
		ret.Results = append(ret.GetResults(), &pb.QueryStateItem{Key: key, Data: value})
	}
	return ret, nil
}

func (s *testDaprServer) DeleteState(ctx context.Context, req *pb.DeleteStateRequest) (*emptypb.Empty, error) {
	delete(s.state, req.GetKey())
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) DeleteBulkState(ctx context.Context, req *pb.DeleteBulkStateRequest) (*emptypb.Empty, error) {
	for _, item := range req.GetStates() {
		delete(s.state, item.GetKey())
	}
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) ExecuteStateTransaction(ctx context.Context, in *pb.ExecuteStateTransactionRequest) (*emptypb.Empty, error) {
	for _, op := range in.GetOperations() {
		item := op.GetRequest()
		switch opType := op.GetOperationType(); opType {
		case "upsert":
			s.state[item.GetKey()] = item.GetValue()
		case "delete":
			delete(s.state, item.GetKey())
		default:
			return &emptypb.Empty{}, fmt.Errorf("invalid operation type: %s", opType)
		}
	}
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (metadata *pb.GetMetadataResponse, err error) {
	resp := &pb.GetMetadataResponse{
		Id:                uuid.NewString(),
		ActiveActorsCount: []*pb.ActiveActorsCount{},
		ExtendedMetadata:  map[string]string{"test_key": "test_value"},
		Subscriptions:     []*pb.PubsubSubscription{},
		HttpEndpoints:     []*pb.MetadataHTTPEndpoint{},
	}
	return resp, nil
}

func (s *testDaprServer) SetMetadata(ctx context.Context, req *pb.SetMetadataRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) PublishEvent(ctx context.Context, req *pb.PublishEventRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) BulkPublishEvent(ctx context.Context, req *pb.BulkPublishRequest) (*pb.BulkPublishResponse, error) {
	return s.bulkPublishEvent(req)
}

// BulkPublishEventAlpha1 mocks the BulkPublishEventAlpha1 API.
func (s *testDaprServer) BulkPublishEventAlpha1(ctx context.Context, req *pb.BulkPublishRequest) (*pb.BulkPublishResponse, error) {
	return s.bulkPublishEvent(req)
}

// It will fail to publish events that start with "fail".
// It will fail the entire request if an event starts with "failall".
func (s *testDaprServer) bulkPublishEvent(req *pb.BulkPublishRequest) (*pb.BulkPublishResponse, error) {
	failedEntries := make([]*pb.BulkPublishResponseFailedEntry, 0)
	for _, entry := range req.GetEntries() {
		if bytes.HasPrefix(entry.GetEvent(), []byte("failall")) {
			// fail the entire request
			return nil, errors.New("failed to publish events")
		} else if bytes.HasPrefix(entry.GetEvent(), []byte("fail")) {
			// fail this entry
			failedEntries = append(failedEntries, &pb.BulkPublishResponseFailedEntry{
				EntryId: entry.GetEntryId(),
				Error:   "failed to publish events",
			})
		}
	}
	return &pb.BulkPublishResponse{FailedEntries: failedEntries}, nil
}

func (s *testDaprServer) InvokeBinding(ctx context.Context, req *pb.InvokeBindingRequest) (*pb.InvokeBindingResponse, error) {
	if req.GetData() == nil {
		return &pb.InvokeBindingResponse{
			Data:     []byte("test"),
			Metadata: map[string]string{"k1": "v1", "k2": "v2"},
		}, nil
	}
	return &pb.InvokeBindingResponse{
		Data:     req.GetData(),
		Metadata: req.GetMetadata(),
	}, nil
}

func (s *testDaprServer) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	d := make(map[string]string)
	d["test"] = "value"
	return &pb.GetSecretResponse{
		Data: d,
	}, nil
}

func (s *testDaprServer) GetBulkSecret(ctx context.Context, req *pb.GetBulkSecretRequest) (*pb.GetBulkSecretResponse, error) {
	d := make(map[string]*pb.SecretResponse)
	d["test"] = &pb.SecretResponse{
		Secrets: map[string]string{
			"test": "value",
		},
	}
	return &pb.GetBulkSecretResponse{
		Data: d,
	}, nil
}

func (s *testDaprServer) RegisterActorReminder(ctx context.Context, req *pb.RegisterActorReminderRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) UnregisterActorReminder(ctx context.Context, req *pb.UnregisterActorReminderRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) InvokeActor(context.Context, *pb.InvokeActorRequest) (*pb.InvokeActorResponse, error) {
	return &pb.InvokeActorResponse{
		Data: []byte("mockValue"),
	}, nil
}

func (s *testDaprServer) RegisterActorTimer(context.Context, *pb.RegisterActorTimerRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) UnregisterActorTimer(context.Context, *pb.UnregisterActorTimerRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) Shutdown(ctx context.Context, req *pb.ShutdownRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) GetConfiguration(ctx context.Context, in *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error) {
	if in.GetStoreName() == "" {
		return &pb.GetConfigurationResponse{}, errors.New("store name notfound")
	}
	items := make(map[string]*commonv1pb.ConfigurationItem)
	for _, v := range in.GetKeys() {
		items[v] = &commonv1pb.ConfigurationItem{
			Value: v + valueSuffix,
		}
	}
	return &pb.GetConfigurationResponse{
		Items: items,
	}, nil
}

func (s *testDaprServer) SubscribeConfiguration(in *pb.SubscribeConfigurationRequest, server pb.Dapr_SubscribeConfigurationServer) error {
	stopCh := make(chan struct{})
	id, _ := uuid.NewUUID()
	s.configurationSubscriptionIDMapLoc.Lock()
	s.configurationSubscriptionID[id.String()] = stopCh
	s.configurationSubscriptionIDMapLoc.Unlock()

	// Send subscription ID in the first response.
	if err := server.Send(&pb.SubscribeConfigurationResponse{
		Id: id.String(),
	}); err != nil {
		return err
	}

	for range 5 {
		select {
		case <-stopCh:
			return nil
		default:
		}
		items := make(map[string]*commonv1pb.ConfigurationItem)
		for _, v := range in.GetKeys() {
			items[v] = &commonv1pb.ConfigurationItem{
				Value: v + valueSuffix,
			}
		}
		if err := server.Send(&pb.SubscribeConfigurationResponse{
			Id:    id.String(),
			Items: items,
		}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (s *testDaprServer) UnsubscribeConfiguration(ctx context.Context, in *pb.UnsubscribeConfigurationRequest) (*pb.UnsubscribeConfigurationResponse, error) {
	s.configurationSubscriptionIDMapLoc.Lock()
	defer s.configurationSubscriptionIDMapLoc.Unlock()
	ch, ok := s.configurationSubscriptionID[in.GetId()]
	if !ok {
		return &pb.UnsubscribeConfigurationResponse{Ok: true}, nil
	}
	close(ch)
	delete(s.configurationSubscriptionID, in.GetId())
	return &pb.UnsubscribeConfigurationResponse{Ok: true}, nil
}

func (s *testDaprServer) ScheduleJobAlpha1(ctx context.Context, in *pb.ScheduleJobRequest) (*pb.ScheduleJobResponse, error) {
	return &pb.ScheduleJobResponse{}, nil
}

func (s *testDaprServer) GetJobAlpha1(ctx context.Context, in *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	var (
		schedule          = "@every 10s"
		dueTime           = "10s"
		repeats    uint32 = 4
		ttl               = "10s"
		maxRetries uint32 = 4
	)
	return &pb.GetJobResponse{
		Job: &pb.Job{
			Name:     "name",
			Schedule: &schedule,
			Repeats:  &repeats,
			DueTime:  &dueTime,
			Ttl:      &ttl,
			Data:     nil,
			FailurePolicy: &commonv1pb.JobFailurePolicy{
				Policy: &commonv1pb.JobFailurePolicy_Constant{
					Constant: &commonv1pb.JobFailurePolicyConstant{
						MaxRetries: &maxRetries,
						Interval:   &durationpb.Duration{Seconds: 10},
					},
				},
			},
		},
	}, nil
}

func (s *testDaprServer) DeleteJobAlpha1(ctx context.Context, in *pb.DeleteJobRequest) (*pb.DeleteJobResponse, error) {
	return &pb.DeleteJobResponse{}, nil
}

// TODO: remove in 1.17
//
//nolint:staticcheck
func (s *testDaprServer) ConverseAlpha1(ctx context.Context, in *pb.ConversationRequest) (*pb.ConversationResponse,
	error,
) {
	return &pb.ConversationResponse{}, nil
}

func (s *testDaprServer) ConverseAlpha2(ctx context.Context, in *pb.ConversationRequestAlpha2) (*pb.
	ConversationResponseAlpha2,
	error,
) {
	return &pb.ConversationResponseAlpha2{}, nil
}

func TestGrpcClient(t *testing.T) {
	protoClient := pb.NewDaprClient(nil)
	client := &GRPCClient{protoClient: protoClient}
	assert.Equal(t, protoClient, client.GrpcClient())
}
