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
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/actor/api"
	actorErr "github.com/dapr/go-sdk/actor/error"
)

const testActorEventsTimeout = 5 * time.Second

func TestSubscribeActorEventsValidation(t *testing.T) {
	_, client := newTestActorEventsServer(t)

	t.Run("nil handler", func(t *testing.T) {
		_, err := client.SubscribeActorEvents(t.Context(), nil, ActorEventSubscriptionOptions{})
		require.Error(t, err)
	})

	t.Run("no registered actor types", func(t *testing.T) {
		handler := &testActorEventHandler{}
		_, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
		require.Error(t, err)
	})
}

func TestSubscribeActorEventsInitialRequest(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	drain := true
	maxStackDepth := int32(16)
	handler := &testActorEventHandler{actorTypes: []string{"testActorType", "otherActorType"}}

	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{
		ActorIdleTimeout:        time.Hour,
		DrainOngoingCallTimeout: 30 * time.Second,
		DrainRebalancedActors:   &drain,
		Reentrancy: &ActorReentrancyConfig{
			Enabled:       true,
			MaxStackDepth: &maxStackDepth,
		},
		EntitiesConfig: []ActorEntityConfig{{
			Entities:         []string{"otherActorType"},
			ActorIdleTimeout: time.Minute,
		}},
	})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	initial := recvWithTimeout(t, srv.initialRequests)
	assert.Equal(t, []string{"testActorType", "otherActorType"}, initial.GetEntities())
	assert.Equal(t, time.Hour, initial.GetActorIdleTimeout().AsDuration())
	assert.Equal(t, 30*time.Second, initial.GetDrainOngoingCallTimeout().AsDuration())
	assert.True(t, initial.GetDrainRebalancedActors())
	require.NotNil(t, initial.GetReentrancy())
	assert.True(t, initial.GetReentrancy().GetEnabled())
	assert.Equal(t, int32(16), initial.GetReentrancy().GetMaxStackDepth())
	require.Len(t, initial.GetEntitiesConfig(), 1)
	assert.Equal(t, []string{"otherActorType"}, initial.GetEntitiesConfig()[0].GetEntities())
	assert.Equal(t, time.Minute, initial.GetEntitiesConfig()[0].GetActorIdleTimeout().AsDuration())
	assert.Nil(t, initial.GetEntitiesConfig()[0].GetReentrancy())
}

func TestSubscribeActorEventsInitialRequestDefaults(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}

	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	initial := recvWithTimeout(t, srv.initialRequests)
	assert.Equal(t, []string{"testActorType"}, initial.GetEntities())
	assert.Nil(t, initial.GetActorIdleTimeout())
	assert.Nil(t, initial.GetDrainOngoingCallTimeout())
	assert.Nil(t, initial.DrainRebalancedActors)
	assert.Nil(t, initial.GetReentrancy())
	assert.Empty(t, initial.GetEntitiesConfig())
}

func TestSubscribeActorEventsInvoke(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}
	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	session := recvWithTimeout(t, srv.sessions)

	t.Run("success", func(t *testing.T) {
		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
				InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
					Id:        "1",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Method:    "GetUser",
					Data:      []byte(`{"name":"test"}`),
					Metadata: map[string]string{
						"content-type":          "application/json",
						"Dapr-Reentrancy-Id":    "reentrancy-id-1",
						"X-Custom-Header-Value": "custom",
					},
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		invokeResp := resp.GetInvokeResponse()
		require.NotNil(t, invokeResp)
		assert.Equal(t, "1", invokeResp.GetId())
		assert.Equal(t, []byte(`{"name":"test"}`), invokeResp.GetData())
		assert.Equal(t, "application/json", invokeResp.GetMetadata()["content-type"])

		calls := handler.invokeCalls()
		require.Len(t, calls, 1)
		assert.Equal(t, "testActorType", calls[0].actorType)
		assert.Equal(t, "myactor", calls[0].actorID)
		assert.Equal(t, "GetUser", calls[0].method)
		assert.Equal(t, []byte(`{"name":"test"}`), calls[0].payload)

		// The reentrancy id from the callback metadata must be available on
		// the handler context for propagation on outbound actor invocations.
		id, ok := reentrancyIDFromContext(calls[0].ctx)
		assert.True(t, ok)
		assert.Equal(t, "reentrancy-id-1", id)
	})

	t.Run("response echoes a non-json content-type", func(t *testing.T) {
		// The content-type must reflect the caller's, not a hardcoded JSON:
		// the actor runtime supports non-JSON serializers.
		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
				InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
					Id:        "ct",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Method:    "GetUser",
					Metadata:  map[string]string{"content-type": "application/x-yaml"},
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		require.NotNil(t, resp.GetInvokeResponse())
		assert.Equal(t, "application/x-yaml", resp.GetInvokeResponse().GetMetadata()["content-type"])
	})

	t.Run("response omits content-type when the caller sent none", func(t *testing.T) {
		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
				InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
					Id:        "no-ct",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Method:    "GetUser",
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		require.NotNil(t, resp.GetInvokeResponse())
		assert.Empty(t, resp.GetInvokeResponse().GetMetadata())
	})

	t.Run("method not found", func(t *testing.T) {
		handler.setInvokeErr(actorErr.ErrActorMethodNoFound)

		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
				InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
					Id:        "2",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Method:    "Missing",
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		failed := resp.GetRequestFailed()
		require.NotNil(t, failed)
		assert.Equal(t, "2", failed.GetId())
		assert.Equal(t, uint32(codes.NotFound), failed.GetCode())
		assert.NotEmpty(t, failed.GetMessage())
	})

	t.Run("invocation failed", func(t *testing.T) {
		handler.setInvokeErr(actorErr.ErrActorInvokeFailed)

		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
				InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
					Id:        "3",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Method:    "GetUser",
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		failed := resp.GetRequestFailed()
		require.NotNil(t, failed)
		assert.Equal(t, "3", failed.GetId())
		assert.Equal(t, uint32(codes.Internal), failed.GetCode())
	})
}

func TestSubscribeActorEventsReminder(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}
	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	session := recvWithTimeout(t, srv.sessions)

	t.Run("success", func(t *testing.T) {
		// The sidecar stores reminder data registered through
		// RegisterActorReminder as its JSON representation, wrapped in a
		// BytesValue Any on the stream.
		registered := []byte(`"reminderdata"`)
		jsonData, err := json.Marshal(registered)
		require.NoError(t, err)
		data, err := anypb.New(wrapperspb.Bytes(jsonData))
		require.NoError(t, err)

		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_ReminderRequest{
				ReminderRequest: &pb.SubscribeActorEventsResponseReminderRequestAlpha1{
					Id:        "10",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Name:      "myreminder",
					DueTime:   "5s",
					Period:    "10s",
					Data:      data,
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		reminderResp := resp.GetReminderResponse()
		require.NotNil(t, reminderResp)
		assert.Equal(t, "10", reminderResp.GetId())
		assert.False(t, reminderResp.GetCancel())

		calls := handler.reminderCalls()
		require.Len(t, calls, 1)
		assert.Equal(t, "testActorType", calls[0].actorType)
		assert.Equal(t, "myactor", calls[0].actorID)
		assert.Equal(t, "myreminder", calls[0].name)

		// The actor runtime must receive the same params shape as the HTTP
		// reminder callback, round-tripping the registered data unchanged.
		var params api.ActorReminderParams
		require.NoError(t, json.Unmarshal(calls[0].params, &params))
		assert.Equal(t, registered, params.Data)
		assert.Equal(t, "5s", params.DueTime)
		assert.Equal(t, "10s", params.Period)
	})

	t.Run("reminder callee not implemented", func(t *testing.T) {
		handler.setReminderErr(actorErr.ErrReminderFuncUndefined)

		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_ReminderRequest{
				ReminderRequest: &pb.SubscribeActorEventsResponseReminderRequestAlpha1{
					Id:        "11",
					ActorType: "testActorType",
					ActorId:   "myactor",
					Name:      "myreminder",
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		failed := resp.GetRequestFailed()
		require.NotNil(t, failed)
		assert.Equal(t, "11", failed.GetId())
		assert.Equal(t, uint32(codes.NotFound), failed.GetCode())
	})
}

func TestSubscribeActorEventsTimer(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}
	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	session := recvWithTimeout(t, srv.sessions)

	// The sidecar stores timer data registered through RegisterActorTimer as
	// its JSON representation, wrapped in a BytesValue Any on the stream.
	registered := []byte(`"timerdata"`)
	jsonData, err := json.Marshal(registered)
	require.NoError(t, err)
	data, err := anypb.New(wrapperspb.Bytes(jsonData))
	require.NoError(t, err)

	session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
		ResponseType: &pb.SubscribeActorEventsResponseAlpha1_TimerRequest{
			TimerRequest: &pb.SubscribeActorEventsResponseTimerRequestAlpha1{
				Id:        "20",
				ActorType: "testActorType",
				ActorId:   "myactor",
				Name:      "mytimer",
				DueTime:   "5s",
				Period:    "10s",
				Callback:  "Invoke",
				Data:      data,
			},
		},
	})

	resp := recvWithTimeout(t, session.responses)
	timerResp := resp.GetTimerResponse()
	require.NotNil(t, timerResp)
	assert.Equal(t, "20", timerResp.GetId())
	assert.False(t, timerResp.GetCancel())

	calls := handler.timerCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "mytimer", calls[0].name)

	// The actor runtime must receive the same params shape as the HTTP timer
	// callback, round-tripping the registered data unchanged.
	var params api.ActorTimerParam
	require.NoError(t, json.Unmarshal(calls[0].params, &params))
	assert.Equal(t, "Invoke", params.CallBack)
	assert.Equal(t, registered, params.Data)
	assert.Equal(t, "5s", params.DueTime)
	assert.Equal(t, "10s", params.Period)
}

func TestSubscribeActorEventsDeactivate(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}
	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	session := recvWithTimeout(t, srv.sessions)

	t.Run("success", func(t *testing.T) {
		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_DeactivateRequest{
				DeactivateRequest: &pb.SubscribeActorEventsResponseDeactivateRequestAlpha1{
					Id:        "30",
					ActorType: "testActorType",
					ActorId:   "myactor",
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		deactivateResp := resp.GetDeactivateResponse()
		require.NotNil(t, deactivateResp)
		assert.Equal(t, "30", deactivateResp.GetId())

		calls := handler.deactivateCalls()
		require.Len(t, calls, 1)
		assert.Equal(t, "testActorType", calls[0].actorType)
		assert.Equal(t, "myactor", calls[0].actorID)
	})

	t.Run("actor not found", func(t *testing.T) {
		handler.setDeactivateErr(actorErr.ErrActorIDNotFound)

		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_DeactivateRequest{
				DeactivateRequest: &pb.SubscribeActorEventsResponseDeactivateRequestAlpha1{
					Id:        "31",
					ActorType: "testActorType",
					ActorId:   "missing",
				},
			},
		})

		resp := recvWithTimeout(t, session.responses)
		failed := resp.GetRequestFailed()
		require.NotNil(t, failed)
		assert.Equal(t, "31", failed.GetId())
		assert.Equal(t, uint32(codes.NotFound), failed.GetCode())
	})
}

func TestSubscribeActorEventsConcurrentDispatch(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	const numCalls = 50

	// Block all invocations until every callback has been dispatched, proving
	// the read loop keeps draining while handlers are in flight and that
	// concurrent sends are serialized.
	unblock := make(chan struct{})
	handler := &testActorEventHandler{
		actorTypes: []string{"testActorType"},
		invokeFn: func(ctx context.Context, actorType, actorID, method string, payload []byte) ([]byte, actorErr.ActorErr) {
			<-unblock
			return payload, actorErr.Success
		},
	}

	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	session := recvWithTimeout(t, srv.sessions)

	for i := range numCalls {
		session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
			ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
				InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
					Id:        strconv.Itoa(i),
					ActorType: "testActorType",
					ActorId:   "myactor",
					Method:    "Echo",
					Data:      []byte(fmt.Sprintf(`"%d"`, i)),
				},
			},
		})
	}
	close(unblock)

	ids := make(map[string]bool, numCalls)
	for range numCalls {
		resp := recvWithTimeout(t, session.responses)
		invokeResp := resp.GetInvokeResponse()
		require.NotNil(t, invokeResp)
		ids[invokeResp.GetId()] = true
		// Response data is correlated with the request id.
		assert.Equal(t, fmt.Sprintf(`"%s"`, invokeResp.GetId()), string(invokeResp.GetData()))
	}
	assert.Len(t, ids, numCalls)
}

func TestSubscribeActorEventsReconnect(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}
	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	recvWithTimeout(t, srv.initialRequests)
	session := recvWithTimeout(t, srv.sessions)

	// Kill the stream server-side; the subscription must reconnect and
	// re-send the initial registration.
	session.kill(status.Error(codes.Unavailable, "sidecar restarting"))

	initial := recvWithTimeout(t, srv.initialRequests)
	assert.Equal(t, []string{"testActorType"}, initial.GetEntities())
	newSession := recvWithTimeout(t, srv.sessions)

	// The new stream is fully functional.
	newSession.push(t, &pb.SubscribeActorEventsResponseAlpha1{
		ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
			InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
				Id:        "1",
				ActorType: "testActorType",
				ActorId:   "myactor",
				Method:    "GetUser",
			},
		},
	})

	resp := recvWithTimeout(t, newSession.responses)
	require.NotNil(t, resp.GetInvokeResponse())
	assert.Equal(t, "1", resp.GetInvokeResponse().GetId())
}

// TestSubscribeActorEventsResponseDroppedAfterReconnect verifies that a
// callback still running when the stream reconnects has its response dropped
// rather than sent on the new stream, where the request id would not
// correlate with any pending request.
func TestSubscribeActorEventsResponseDroppedAfterReconnect(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	entered := make(chan struct{})
	release := make(chan struct{})
	handler := &testActorEventHandler{
		actorTypes: []string{"testActorType"},
		invokeFn: func(ctx context.Context, actorType, actorID, method string, payload []byte) ([]byte, actorErr.ActorErr) {
			if method == "Blocking" {
				close(entered)
				<-release // hold the response until after the reconnect
			}
			return payload, actorErr.Success
		},
	}

	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)
	defer stop() //nolint:errcheck

	recvWithTimeout(t, srv.initialRequests)
	session := recvWithTimeout(t, srv.sessions)

	// Deliver a callback that blocks inside the handler, capturing the
	// original stream.
	session.push(t, &pb.SubscribeActorEventsResponseAlpha1{
		ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
			InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
				Id: "stale", ActorType: "testActorType", ActorId: "myactor", Method: "Blocking",
			},
		},
	})
	<-entered

	// Reconnect while the handler is still running.
	session.kill(status.Error(codes.Unavailable, "sidecar restarting"))
	recvWithTimeout(t, srv.initialRequests)
	newSession := recvWithTimeout(t, srv.sessions)

	// Let the stale handler finish: its response targets the old stream and
	// must be dropped, never reaching the new stream.
	close(release)

	// A fresh callback on the new stream still works.
	newSession.push(t, &pb.SubscribeActorEventsResponseAlpha1{
		ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InvokeRequest{
			InvokeRequest: &pb.SubscribeActorEventsResponseInvokeRequestAlpha1{
				Id: "fresh", ActorType: "testActorType", ActorId: "myactor", Method: "Echo",
			},
		},
	})

	resp := recvWithTimeout(t, newSession.responses)
	require.NotNil(t, resp.GetInvokeResponse())
	assert.Equal(t, "fresh", resp.GetInvokeResponse().GetId(),
		"the new stream must only receive the fresh response, not the stale one")

	// No further (stale) response should arrive on the new stream.
	select {
	case extra := <-newSession.responses:
		t.Fatalf("unexpected extra response on reconnected stream: %v", extra)
	case <-time.After(500 * time.Millisecond):
	}
}

func TestSubscribeActorEventsStop(t *testing.T) {
	srv, client := newTestActorEventsServer(t)

	handler := &testActorEventHandler{actorTypes: []string{"testActorType"}}
	stop, err := client.SubscribeActorEvents(t.Context(), handler, ActorEventSubscriptionOptions{})
	require.NoError(t, err)

	recvWithTimeout(t, srv.initialRequests)
	recvWithTimeout(t, srv.sessions)

	require.NoError(t, stop())
	require.Error(t, stop(), "second stop must report the subscription is already closed")

	// A stopped subscription must not reconnect.
	select {
	case <-srv.initialRequests:
		t.Fatal("subscription reconnected after being stopped")
	case <-time.After(500 * time.Millisecond):
	}
}

func Test_actorErrToStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  actorErr.ActorErr
		want codes.Code
	}{
		{"actor type not found", actorErr.ErrActorTypeNotFound, codes.NotFound},
		{"actor method not found", actorErr.ErrActorMethodNoFound, codes.NotFound},
		{"actor id not found", actorErr.ErrActorIDNotFound, codes.NotFound},
		{"reminder func undefined", actorErr.ErrReminderFuncUndefined, codes.NotFound},
		{"reminder params invalid", actorErr.ErrRemindersParamsInvalid, codes.InvalidArgument},
		{"timer params invalid", actorErr.ErrTimerParamsInvalid, codes.InvalidArgument},
		{"invoke failed", actorErr.ErrActorInvokeFailed, codes.Internal},
		{"save state failed", actorErr.ErrSaveStateFailed, codes.Internal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, actorErrToStatusCode(tt.err))
		})
	}
}

func Test_anyToRawData(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, anyToRawData(nil))
	})

	t.Run("wrapped bytes", func(t *testing.T) {
		data, err := anypb.New(wrapperspb.Bytes([]byte(`{"key":"value"}`)))
		require.NoError(t, err)
		assert.Equal(t, []byte(`{"key":"value"}`), anyToRawData(data))
	})

	t.Run("typed message", func(t *testing.T) {
		data, err := anypb.New(wrapperspb.String("hello"))
		require.NoError(t, err)
		assert.Equal(t, []byte(`"hello"`), anyToRawData(data))
	})

	t.Run("unknown type falls back to a valid JSON string", func(t *testing.T) {
		data := &anypb.Any{
			TypeUrl: "type.googleapis.com/unknown.Type",
			Value:   []byte("raw"),
		}
		got := anyToRawData(data)

		// The fallback must be valid JSON (it is embedded as json.RawMessage in
		// the reminder/timer params) and must round-trip back to the raw bytes,
		// matching how api.ActorReminderParams.Data ([]byte) decodes it.
		assert.True(t, json.Valid(got), "fallback must be valid JSON: %s", got)
		var decoded []byte
		require.NoError(t, json.Unmarshal(got, &decoded))
		assert.Equal(t, []byte("raw"), decoded)
	})
}

func Test_reentrancyIDFromContext(t *testing.T) {
	t.Run("no metadata", func(t *testing.T) {
		_, ok := reentrancyIDFromContext(t.Context())
		assert.False(t, ok)
	})

	t.Run("metadata without reentrancy id", func(t *testing.T) {
		ctx := withActorCallbackMetadata(t.Context(), map[string]string{"content-type": "application/json"})
		_, ok := reentrancyIDFromContext(ctx)
		assert.False(t, ok)
	})

	t.Run("reentrancy id", func(t *testing.T) {
		ctx := withActorCallbackMetadata(t.Context(), map[string]string{"Dapr-Reentrancy-Id": "id-1"})
		id, ok := reentrancyIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "id-1", id)
	})

	t.Run("case insensitive", func(t *testing.T) {
		ctx := withActorCallbackMetadata(t.Context(), map[string]string{"dapr-reentrancy-id": "id-2"})
		id, ok := reentrancyIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "id-2", id)
	})
}

// newTestActorEventsServer starts a dedicated in-memory Dapr server
// implementing SubscribeActorEventsAlpha1 and returns it along with a client
// connected to it.
func newTestActorEventsServer(t *testing.T) (*fakeActorEventsServer, Client) {
	t.Helper()

	srv := &fakeActorEventsServer{
		initialRequests: make(chan *pb.SubscribeActorEventsRequestInitialAlpha1, 16),
		sessions:        make(chan *actorEventsTestSession, 16),
	}

	s := grpc.NewServer()
	pb.RegisterDaprServer(s, srv)

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

	return srv, NewClientWithConnection(conn)
}

type fakeActorEventsServer struct {
	pb.UnimplementedDaprServer

	initialRequests chan *pb.SubscribeActorEventsRequestInitialAlpha1
	sessions        chan *actorEventsTestSession
}

// actorEventsTestSession represents one established actor events stream. The
// test pushes callbacks to the app with push and reads the app's responses
// from responses. kill makes the server side fail the stream.
type actorEventsTestSession struct {
	stream    pb.Dapr_SubscribeActorEventsAlpha1Server
	responses chan *pb.SubscribeActorEventsRequestAlpha1
	killCh    chan error
	recvDone  chan struct{}
}

func (s *fakeActorEventsServer) SubscribeActorEventsAlpha1(stream pb.Dapr_SubscribeActorEventsAlpha1Server) error {
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	initial := first.GetInitialRequest()
	if initial == nil {
		return status.Error(codes.InvalidArgument, "first message must be the initial request")
	}
	s.initialRequests <- initial

	if err := stream.Send(&pb.SubscribeActorEventsResponseAlpha1{
		ResponseType: &pb.SubscribeActorEventsResponseAlpha1_InitialResponse{
			InitialResponse: &pb.SubscribeActorEventsResponseInitialAlpha1{},
		},
	}); err != nil {
		return err
	}

	session := &actorEventsTestSession{
		stream:    stream,
		responses: make(chan *pb.SubscribeActorEventsRequestAlpha1, 64),
		killCh:    make(chan error, 1),
		recvDone:  make(chan struct{}),
	}
	s.sessions <- session

	go func() {
		defer close(session.recvDone)
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}
			session.responses <- msg
		}
	}()

	select {
	case err := <-session.killCh:
		return err
	case <-session.recvDone:
		return nil
	case <-stream.Context().Done():
		return nil
	}
}

func (s *actorEventsTestSession) push(t *testing.T, msg *pb.SubscribeActorEventsResponseAlpha1) {
	t.Helper()
	require.NoError(t, s.stream.Send(msg))
}

func (s *actorEventsTestSession) kill(err error) {
	s.killCh <- err
}

func recvWithTimeout[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(testActorEventsTimeout):
		t.Fatal("timed out waiting for message")
		panic("unreachable")
	}
}

type invokeCall struct {
	ctx       context.Context
	actorType string
	actorID   string
	method    string
	payload   []byte
}

type reminderCall struct {
	actorType string
	actorID   string
	name      string
	params    []byte
}

type deactivateCall struct {
	actorType string
	actorID   string
}

// testActorEventHandler is a configurable ActorEventHandler recording every
// dispatched callback.
type testActorEventHandler struct {
	actorTypes []string

	invokeFn func(ctx context.Context, actorType, actorID, method string, payload []byte) ([]byte, actorErr.ActorErr)

	lock          sync.Mutex
	invokes       []invokeCall
	reminders     []reminderCall
	timers        []reminderCall
	deactivations []deactivateCall
	invokeErr     actorErr.ActorErr
	reminderErr   actorErr.ActorErr
	timerErr      actorErr.ActorErr
	deactivateErr actorErr.ActorErr
}

func (h *testActorEventHandler) RegisteredActorTypes() []string {
	return h.actorTypes
}

func (h *testActorEventHandler) InvokeActorMethod(ctx context.Context, actorType, actorID, method string, payload []byte) ([]byte, actorErr.ActorErr) {
	h.lock.Lock()
	h.invokes = append(h.invokes, invokeCall{ctx: ctx, actorType: actorType, actorID: actorID, method: method, payload: payload})
	invokeErr := h.invokeErr
	invokeFn := h.invokeFn
	h.lock.Unlock()

	if invokeFn != nil {
		return invokeFn(ctx, actorType, actorID, method, payload)
	}
	if invokeErr != actorErr.Success {
		return nil, invokeErr
	}
	return payload, actorErr.Success
}

func (h *testActorEventHandler) InvokeReminder(ctx context.Context, actorType, actorID, reminderName string, params []byte) actorErr.ActorErr {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.reminders = append(h.reminders, reminderCall{actorType: actorType, actorID: actorID, name: reminderName, params: params})
	return h.reminderErr
}

func (h *testActorEventHandler) InvokeTimer(ctx context.Context, actorType, actorID, timerName string, params []byte) actorErr.ActorErr {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.timers = append(h.timers, reminderCall{actorType: actorType, actorID: actorID, name: timerName, params: params})
	return h.timerErr
}

func (h *testActorEventHandler) Deactivate(ctx context.Context, actorType, actorID string) actorErr.ActorErr {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.deactivations = append(h.deactivations, deactivateCall{actorType: actorType, actorID: actorID})
	return h.deactivateErr
}

func (h *testActorEventHandler) invokeCalls() []invokeCall {
	h.lock.Lock()
	defer h.lock.Unlock()
	return append([]invokeCall(nil), h.invokes...)
}

func (h *testActorEventHandler) reminderCalls() []reminderCall {
	h.lock.Lock()
	defer h.lock.Unlock()
	return append([]reminderCall(nil), h.reminders...)
}

func (h *testActorEventHandler) timerCalls() []reminderCall {
	h.lock.Lock()
	defer h.lock.Unlock()
	return append([]reminderCall(nil), h.timers...)
}

func (h *testActorEventHandler) deactivateCalls() []deactivateCall {
	h.lock.Lock()
	defer h.lock.Unlock()
	return append([]deactivateCall(nil), h.deactivations...)
}

func (h *testActorEventHandler) setInvokeErr(err actorErr.ActorErr) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.invokeErr = err
}

func (h *testActorEventHandler) setReminderErr(err actorErr.ActorErr) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.reminderErr = err
}

func (h *testActorEventHandler) setDeactivateErr(err actorErr.ActorErr) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.deactivateErr = err
}
