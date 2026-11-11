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
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/kit/events/loop"
)

const (
	// reentrancyIDMetadataKey is the metadata key under which daprd sends the
	// reentrancy id on actor invoke callbacks. It is propagated back on
	// outbound actor invocations performed from within the callback context.
	reentrancyIDMetadataKey = "Dapr-Reentrancy-Id"

	// contentTypeMetadataKey carries the payload content type on actor
	// invoke callbacks and responses.
	contentTypeMetadataKey = "content-type"

	actorEventsReconnectInitialBackoff = 200 * time.Millisecond
	actorEventsReconnectMaxBackoff     = 5 * time.Second

	// actorEventsSendBuffer is the per-segment capacity of the outbound send
	// loop's queue. The queue grows by segments under burst, so this only sets
	// how often that happens.
	actorEventsSendBuffer = 64
)

// actorEventsSendLoopFactory builds the single-consumer loop that serializes
// outbound writes to the stream. gRPC forbids concurrent Send on one stream;
// running every write through one loop goroutine serializes them without
// holding a lock across the (potentially blocking) Send. Mirrors the
// loop-per-stream pattern daprd uses for its own actor streams.
var actorEventsSendLoopFactory = loop.New[actorEventOutbound](actorEventsSendBuffer)

// actorEventOutbound is a single write to perform on the actor event stream,
// bound to the stream that produced the originating callback so a response
// left over from a stream that has since reconnected can be dropped. A nil msg
// is the loop's close sentinel.
type actorEventOutbound struct {
	stream pb.Dapr_SubscribeActorEventsAlpha1Client
	msg    *pb.SubscribeActorEventsRequestAlpha1
}

// ActorEventHandler dispatches actor callbacks received over the actor event
// stream to the hosted actor implementations.
// *actor/runtime.ActorRunTimeContext satisfies this interface.
type ActorEventHandler interface {
	// RegisteredActorTypes returns the actor types hosted by this handler.
	RegisteredActorTypes() []string
	// InvokeActorMethod invokes an actor method and returns the serialized
	// response payload.
	InvokeActorMethod(ctx context.Context, actorType, actorID, method string, payload []byte) ([]byte, actorErr.ActorErr)
	// InvokeReminder invokes a reminder on an actor. params is the JSON
	// serialized api.ActorReminderParams.
	InvokeReminder(ctx context.Context, actorType, actorID, reminderName string, params []byte) actorErr.ActorErr
	// InvokeTimer invokes a timer on an actor. params is the JSON serialized
	// api.ActorTimerParam.
	InvokeTimer(ctx context.Context, actorType, actorID, timerName string, params []byte) actorErr.ActorErr
	// Deactivate deactivates an actor instance.
	Deactivate(ctx context.Context, actorType, actorID string) actorErr.ActorErr
}

// ActorReentrancyConfig configures actor reentrancy for the registered actor
// types. A nil MaxStackDepth means Dapr's default is used.
type ActorReentrancyConfig struct {
	Enabled       bool
	MaxStackDepth *int32
}

// ActorEntityConfig overrides the default actor configuration for a subset of
// the registered actor types. Zero values mean the defaults from
// ActorEventSubscriptionOptions (or Dapr's own defaults) apply.
type ActorEntityConfig struct {
	Entities                []string
	ActorIdleTimeout        time.Duration
	DrainOngoingCallTimeout time.Duration
	DrainRebalancedActors   *bool
	Reentrancy              *ActorReentrancyConfig
}

// ActorEventSubscriptionOptions configures the actor registration sent to the
// Dapr sidecar when subscribing to actor events. Zero values mean Dapr's
// defaults are used.
type ActorEventSubscriptionOptions struct {
	// ActorIdleTimeout is the time an actor can stay idle before it is
	// deactivated.
	ActorIdleTimeout time.Duration
	// DrainOngoingCallTimeout is how long Dapr waits for ongoing actor calls
	// to complete when draining rebalanced actors.
	DrainOngoingCallTimeout time.Duration
	// DrainRebalancedActors enables draining of actors that are rebalanced to
	// another host. A nil value means Dapr's default is used.
	DrainRebalancedActors *bool
	// Reentrancy configures actor reentrancy for the registered actor types.
	Reentrancy *ActorReentrancyConfig
	// EntitiesConfig overrides the defaults above for specific actor types.
	EntitiesConfig []ActorEntityConfig
}

// actorEventSubscription hosts actor types over the app-initiated
// SubscribeActorEventsAlpha1 stream, dispatching callbacks pushed by the
// sidecar to the ActorEventHandler and sending the correlated responses back.
type actorEventSubscription struct {
	// ctx is an internal context derived from the caller's; cancel tears it
	// down so a Recv blocked in run() unblocks promptly on Close.
	ctx    context.Context
	cancel context.CancelFunc

	handler ActorEventHandler

	// sendLoop serializes all writes to the stream on a single goroutine, so a
	// blocking Send never stalls the read loop or the dispatch goroutines.
	sendLoop loop.Interface[actorEventOutbound]

	// lock guards the stream pointer, which reconnect replaces.
	lock   sync.Mutex
	stream pb.Dapr_SubscribeActorEventsAlpha1Client
	closed atomic.Bool

	createStream func(ctx context.Context) (pb.Dapr_SubscribeActorEventsAlpha1Client, error)
}

// SubscribeActorEvents registers the handler's actor types with the Dapr
// sidecar and hosts them over a bidirectional event stream: actor method
// invocations, reminders, timers, and deactivations are delivered through the
// stream, so the application does not need to expose a server port for actor
// callbacks. The stream automatically reconnects and re-registers on
// transient failures.
//
// Hosting stops when the returned stop function is called or ctx is
// cancelled. Actor types must be registered on the handler (e.g. with
// runtime.ActorRunTimeContext.RegisterActorFactory) before subscribing.
//
// This API is currently in Alpha.
func (c *GRPCClient) SubscribeActorEvents(ctx context.Context, handler ActorEventHandler, opts ActorEventSubscriptionOptions) (func() error, error) {
	if handler == nil {
		return nil, errors.New("actor event handler required")
	}
	if len(handler.RegisteredActorTypes()) == 0 {
		return nil, errors.New("actor event handler has no registered actor types")
	}

	createStream := func(ctx context.Context) (pb.Dapr_SubscribeActorEventsAlpha1Client, error) {
		// Resolve the actor types on every (re)connect so types registered
		// after a disconnect are picked up on re-registration.
		return c.subscribeActorEventsInitialRequest(ctx, handler.RegisteredActorTypes(), opts)
	}

	// Derive an internal context so Close can cancel the stream (and any
	// in-flight Recv) without depending on the caller cancelling ctx.
	streamCtx, cancel := context.WithCancel(ctx)

	stream, err := createStream(streamCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	s := &actorEventSubscription{
		ctx:          streamCtx,
		cancel:       cancel,
		handler:      handler,
		stream:       stream,
		createStream: createStream,
	}
	s.sendLoop = actorEventsSendLoopFactory.NewLoop(s)

	go func() { _ = s.sendLoop.Run(streamCtx) }()
	go s.run()

	return s.Close, nil
}

// subscribeActorEventsInitialRequest opens the actor events stream, sends the
// initial registration message and waits for the registration ack.
func (c *GRPCClient) subscribeActorEventsInitialRequest(ctx context.Context, entities []string, opts ActorEventSubscriptionOptions) (pb.Dapr_SubscribeActorEventsAlpha1Client, error) {
	stream, err := c.protoClient.SubscribeActorEventsAlpha1(ctx)
	if err != nil {
		return nil, err
	}

	err = stream.Send(&pb.SubscribeActorEventsRequestAlpha1{
		RequestType: &pb.SubscribeActorEventsRequestAlpha1_InitialRequest{
			InitialRequest: actorEventsInitialRequest(entities, opts),
		},
	})
	if err != nil {
		return nil, errors.Join(err, stream.CloseSend())
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, errors.Join(err, stream.CloseSend())
	}

	if resp.GetInitialResponse() == nil {
		return nil, errors.Join(
			fmt.Errorf("unexpected initial response from server: %v", resp),
			stream.CloseSend(),
		)
	}

	return stream, nil
}

// actorEventsInitialRequest maps the registered actor types and subscription
// options to the wire-level registration message. Unset options are left nil
// so Dapr applies its own defaults.
func actorEventsInitialRequest(entities []string, opts ActorEventSubscriptionOptions) *pb.SubscribeActorEventsRequestInitialAlpha1 {
	req := &pb.SubscribeActorEventsRequestInitialAlpha1{
		Entities:                entities,
		ActorIdleTimeout:        optionalDuration(opts.ActorIdleTimeout),
		DrainOngoingCallTimeout: optionalDuration(opts.DrainOngoingCallTimeout),
		DrainRebalancedActors:   opts.DrainRebalancedActors,
		Reentrancy:              opts.Reentrancy.proto(),
	}
	for _, ec := range opts.EntitiesConfig {
		req.EntitiesConfig = append(req.EntitiesConfig, &pb.ActorEntityConfig{
			Entities:                ec.Entities,
			ActorIdleTimeout:        optionalDuration(ec.ActorIdleTimeout),
			DrainOngoingCallTimeout: optionalDuration(ec.DrainOngoingCallTimeout),
			DrainRebalancedActors:   ec.DrainRebalancedActors,
			Reentrancy:              ec.Reentrancy.proto(),
		})
	}
	return req
}

func (r *ActorReentrancyConfig) proto() *pb.ActorReentrancyConfig {
	if r == nil {
		return nil
	}
	return &pb.ActorReentrancyConfig{
		Enabled:       r.Enabled,
		MaxStackDepth: r.MaxStackDepth,
	}
}

func optionalDuration(d time.Duration) *durationpb.Duration {
	if d == 0 {
		return nil
	}
	return durationpb.New(d)
}

// Close stops hosting actor events. Cancelling the stream context tears down
// the stream (unblocking a Recv in run() and any Send in the loop), then the
// send loop is drained and stopped. No lock is taken, so a blocked Send cannot
// deadlock shutdown.
func (s *actorEventSubscription) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return errors.New("actor event subscription already closed")
	}

	s.cancel()
	s.sendLoop.Close(actorEventOutbound{})
	return nil
}

// run reads callbacks from the stream and dispatches each on its own
// goroutine so the read loop keeps draining. On a transient stream error it
// reconnects with a backoff, re-sending the initial registration.
func (s *actorEventSubscription) run() {
	// Ensure the send loop is stopped when the reader exits, including when the
	// caller cancels ctx without calling the returned stop func. Close is
	// idempotent, so a concurrent stop() is harmless.
	defer s.Close() //nolint:errcheck

	for {
		// Snapshot the current stream under the lock; reconnect may replace it
		// concurrently with the send loop.
		s.lock.Lock()
		stream := s.stream
		s.lock.Unlock()

		msg, err := stream.Recv()
		if err != nil {
			if s.closed.Load() || s.ctx.Err() != nil {
				return
			}

			if st, ok := status.FromError(err); ok && st.Code() == codes.Canceled {
				return
			}

			logger.Printf("Error receiving actor events, reconnecting: %v", err)
			if !s.reconnect() {
				return
			}
			continue
		}

		// Dispatch on the stream that delivered the message so the response is
		// sent back on the same stream, keeping request ids correlated even if
		// a reconnect swaps the stream while the handler runs.
		go s.dispatch(stream, msg)
	}
}

// reconnect re-establishes the actor events stream with an exponential
// backoff, returning false when the subscription is closed or its context is
// cancelled.
func (s *actorEventSubscription) reconnect() bool {
	backoff := actorEventsReconnectInitialBackoff
	for {
		select {
		case <-s.ctx.Done():
			return false
		case <-time.After(backoff):
		}
		if s.closed.Load() {
			return false
		}

		newStream, err := s.createStream(s.ctx)
		if err != nil {
			logger.Printf("Error reconnecting actor events stream, retrying in %s: %v", backoff, err)
			backoff = min(backoff*2, actorEventsReconnectMaxBackoff)
			continue
		}

		s.lock.Lock()
		s.stream = newStream
		s.lock.Unlock()

		return true
	}
}

// dispatch routes a single callback pushed by the sidecar to the handler and
// sends the correlated response on stream, echoing the request id.
func (s *actorEventSubscription) dispatch(stream pb.Dapr_SubscribeActorEventsAlpha1Client, msg *pb.SubscribeActorEventsResponseAlpha1) {
	switch t := msg.GetResponseType().(type) {
	case *pb.SubscribeActorEventsResponseAlpha1_InvokeRequest:
		s.handleInvoke(stream, t.InvokeRequest)
	case *pb.SubscribeActorEventsResponseAlpha1_ReminderRequest:
		s.handleReminder(stream, t.ReminderRequest)
	case *pb.SubscribeActorEventsResponseAlpha1_TimerRequest:
		s.handleTimer(stream, t.TimerRequest)
	case *pb.SubscribeActorEventsResponseAlpha1_DeactivateRequest:
		s.handleDeactivate(stream, t.DeactivateRequest)
	default:
		logger.Printf("Unexpected actor event message type: %T", t)
	}
}

func (s *actorEventSubscription) handleInvoke(stream pb.Dapr_SubscribeActorEventsAlpha1Client, req *pb.SubscribeActorEventsResponseInvokeRequestAlpha1) {
	// Make the callback metadata (content-type, Dapr-Reentrancy-Id, ...)
	// available to the actor code: InvokeActor propagates the reentrancy id
	// from this context on outbound actor invocations.
	ctx := withActorCallbackMetadata(s.ctx, req.GetMetadata())

	data, aerr := s.handler.InvokeActorMethod(ctx, req.GetActorType(), req.GetActorId(), req.GetMethod(), req.GetData())
	if aerr != actorErr.Success {
		s.sendRequestFailed(stream, req.GetId(), aerr)
		return
	}

	// Echo the caller's content-type back on the response. The actor runtime
	// (de)serializes the request and response with the same per-type codec, so
	// the request content-type describes the response payload too. daprd passes
	// this back to the original caller verbatim; leaving it unset when the
	// caller sent none mirrors the HTTP actor callback, which sets no explicit
	// response content-type either.
	var metadata map[string]string
	if ct := contentTypeFromMetadata(req.GetMetadata()); ct != "" {
		metadata = map[string]string{contentTypeMetadataKey: ct}
	}

	s.send(stream, &pb.SubscribeActorEventsRequestAlpha1{
		RequestType: &pb.SubscribeActorEventsRequestAlpha1_InvokeResponse{
			InvokeResponse: &pb.SubscribeActorEventsRequestInvokeResponseAlpha1{
				Id:       req.GetId(),
				Data:     data,
				Metadata: metadata,
			},
		},
	})
}

// contentTypeFromMetadata returns the content-type entry from actor callback
// metadata, matching the key case-insensitively.
func contentTypeFromMetadata(md map[string]string) string {
	for k, v := range md {
		if strings.EqualFold(k, contentTypeMetadataKey) {
			return v
		}
	}
	return ""
}

func (s *actorEventSubscription) handleReminder(stream pb.Dapr_SubscribeActorEventsAlpha1Client, req *pb.SubscribeActorEventsResponseReminderRequestAlpha1) {
	// Build the same JSON shape the sidecar sends on the HTTP reminder
	// callback, which api.ActorReminderParams unmarshals: data carries the
	// JSON representation of the registered payload verbatim.
	params, err := json.Marshal(&struct {
		Data    json.RawMessage `json:"data,omitempty"`
		DueTime string          `json:"dueTime"`
		Period  string          `json:"period"`
	}{
		Data:    anyToRawData(req.GetData()),
		DueTime: req.GetDueTime(),
		Period:  req.GetPeriod(),
	})
	if err != nil {
		s.sendRequestFailed(stream, req.GetId(), actorErr.ErrRemindersParamsInvalid)
		return
	}

	aerr := s.handler.InvokeReminder(s.ctx, req.GetActorType(), req.GetActorId(), req.GetName(), params)
	if aerr != actorErr.Success {
		s.sendRequestFailed(stream, req.GetId(), aerr)
		return
	}

	s.send(stream, &pb.SubscribeActorEventsRequestAlpha1{
		RequestType: &pb.SubscribeActorEventsRequestAlpha1_ReminderResponse{
			ReminderResponse: &pb.SubscribeActorEventsRequestReminderResponseAlpha1{Id: req.GetId()},
		},
	})
}

func (s *actorEventSubscription) handleTimer(stream pb.Dapr_SubscribeActorEventsAlpha1Client, req *pb.SubscribeActorEventsResponseTimerRequestAlpha1) {
	// Build the same JSON shape the sidecar sends on the HTTP timer callback,
	// which api.ActorTimerParam unmarshals: data carries the JSON
	// representation of the registered payload verbatim.
	params, err := json.Marshal(&struct {
		Callback string          `json:"callback"`
		Data     json.RawMessage `json:"data,omitempty"`
		DueTime  string          `json:"dueTime"`
		Period   string          `json:"period"`
	}{
		Callback: req.GetCallback(),
		Data:     anyToRawData(req.GetData()),
		DueTime:  req.GetDueTime(),
		Period:   req.GetPeriod(),
	})
	if err != nil {
		s.sendRequestFailed(stream, req.GetId(), actorErr.ErrTimerParamsInvalid)
		return
	}

	aerr := s.handler.InvokeTimer(s.ctx, req.GetActorType(), req.GetActorId(), req.GetName(), params)
	if aerr != actorErr.Success {
		s.sendRequestFailed(stream, req.GetId(), aerr)
		return
	}

	s.send(stream, &pb.SubscribeActorEventsRequestAlpha1{
		RequestType: &pb.SubscribeActorEventsRequestAlpha1_TimerResponse{
			TimerResponse: &pb.SubscribeActorEventsRequestReminderResponseAlpha1{Id: req.GetId()},
		},
	})
}

func (s *actorEventSubscription) handleDeactivate(stream pb.Dapr_SubscribeActorEventsAlpha1Client, req *pb.SubscribeActorEventsResponseDeactivateRequestAlpha1) {
	aerr := s.handler.Deactivate(s.ctx, req.GetActorType(), req.GetActorId())
	if aerr != actorErr.Success {
		s.sendRequestFailed(stream, req.GetId(), aerr)
		return
	}

	s.send(stream, &pb.SubscribeActorEventsRequestAlpha1{
		RequestType: &pb.SubscribeActorEventsRequestAlpha1_DeactivateResponse{
			DeactivateResponse: &pb.SubscribeActorEventsRequestDeactivateResponseAlpha1{Id: req.GetId()},
		},
	})
}

// sendRequestFailed reports a callback the handler could not process. The
// gRPC status code matters to the sidecar: NotFound marks the failure as
// permanent (non-retryable), e.g. an unknown actor type or method.
func (s *actorEventSubscription) sendRequestFailed(stream pb.Dapr_SubscribeActorEventsAlpha1Client, id string, aerr actorErr.ActorErr) {
	s.send(stream, &pb.SubscribeActorEventsRequestAlpha1{
		RequestType: &pb.SubscribeActorEventsRequestAlpha1_RequestFailed{
			RequestFailed: &pb.SubscribeActorEventsRequestFailedAlpha1{
				Id:      id,
				Code:    uint32(actorErrToStatusCode(aerr)),
				Message: actorErrMessage(aerr),
			},
		},
	})
}

// send queues a write to the stream. Enqueue is non-blocking, so dispatch
// goroutines never block here; the send loop performs the actual write.
func (s *actorEventSubscription) send(stream pb.Dapr_SubscribeActorEventsAlpha1Client, msg *pb.SubscribeActorEventsRequestAlpha1) {
	s.sendLoop.Enqueue(actorEventOutbound{stream: stream, msg: msg})
}

// Handle performs one queued write on the send loop's single goroutine, which
// serializes Send as gRPC requires without holding a lock across it. It writes
// to the specific stream that delivered the callback, dropping the response if
// a reconnect has since replaced the stream. A nil msg is the close sentinel.
func (s *actorEventSubscription) Handle(_ context.Context, out actorEventOutbound) error {
	if out.msg == nil {
		return nil
	}

	s.lock.Lock()
	current := s.stream
	s.lock.Unlock()

	if out.stream != current {
		// A reconnect replaced the stream while this callback ran. The request
		// id is meaningless on the new stream, and the sidecar already failed
		// the in-flight request when the old stream dropped, so drop it.
		logger.Printf("Dropping actor event response on a reconnected stream")
		return nil
	}

	if err := out.stream.Send(out.msg); err != nil && !s.closed.Load() {
		// The response cannot be delivered (e.g. the stream dropped after the
		// callback was dispatched); the sidecar fails the pending request on
		// its side when the stream breaks.
		logger.Printf("Error sending actor event response: %v", err)
	}
	// Never return an error: that would stop the loop and abort all further
	// writes for this subscription.
	return nil
}

// actorErrToStatusCode maps actor runtime errors to the gRPC status codes
// reported on request_failed messages. The mapping mirrors the HTTP actor
// service, where *NotFound errors map to 404 and the rest to 500.
func actorErrToStatusCode(aerr actorErr.ActorErr) codes.Code {
	switch aerr {
	case actorErr.ErrActorTypeNotFound,
		actorErr.ErrActorMethodNoFound,
		actorErr.ErrActorIDNotFound,
		actorErr.ErrReminderFuncUndefined:
		return codes.NotFound
	case actorErr.ErrRemindersParamsInvalid,
		actorErr.ErrTimerParamsInvalid:
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

func actorErrMessage(aerr actorErr.ActorErr) string {
	switch aerr {
	case actorErr.ErrActorTypeNotFound:
		return "actor type not found"
	case actorErr.ErrActorMethodNoFound:
		return "actor method not found"
	case actorErr.ErrActorIDNotFound:
		return "actor id not found"
	case actorErr.ErrReminderFuncUndefined:
		return "actor does not implement reminder callbacks"
	case actorErr.ErrRemindersParamsInvalid:
		return "invalid reminder parameters"
	case actorErr.ErrTimerParamsInvalid:
		return "invalid timer parameters"
	case actorErr.ErrActorMethodSerializeFailed:
		return "failed to serialize actor method payload"
	case actorErr.ErrActorInvokeFailed:
		return "actor method invocation failed"
	case actorErr.ErrActorFactoryNotSet:
		return "actor factory not set"
	case actorErr.ErrSaveStateFailed:
		return "failed to save actor state"
	default:
		return fmt.Sprintf("actor error %d", aerr)
	}
}

// anyToRawData unwraps the reminder/timer payload the sidecar sends as a
// google.protobuf.Any into the JSON value to embed under the params "data"
// field. The returned bytes are a JSON value, NOT raw bytes, so the caller
// must embed them as json.RawMessage; marshalling them as []byte would
// base64-encode them a second time and the actor runtime would decode the
// wrong bytes.
//
// Payloads registered via the actor reminder/timer APIs arrive as a
// BytesValue whose contents are already json.Marshal(registered) — i.e. a
// base64-encoded JSON string — because the sidecar marshals the bytes before
// wrapping them (see RegisterActorReminder/RegisterActorTimer in daprd's
// pkg/api/universal/actors.go). api.ActorReminderParams.Data ([]byte) then
// base64-decodes that string back to the registered bytes. Typed payloads are
// rendered as JSON via protojson. This mirrors how the sidecar serializes
// reminder/timer data for the HTTP callbacks.
func anyToRawData(data *anypb.Any) []byte {
	if data == nil {
		return nil
	}
	msg, err := data.UnmarshalNew()
	if err != nil {
		// Unknown or unregistered payload type: encode the raw bytes as a JSON
		// string so the value stays valid JSON when embedded under params
		// "data". The actor runtime base64-decodes it back to these bytes.
		return rawBytesToJSON(data.GetValue())
	}
	if b, ok := msg.(*wrapperspb.BytesValue); ok {
		return b.GetValue()
	}
	d, err := protojson.Marshal(msg)
	if err != nil {
		return rawBytesToJSON(data.GetValue())
	}
	return d
}

// rawBytesToJSON encodes opaque bytes as a JSON string (base64) so they can be
// embedded as a valid JSON value. json.Marshal of a []byte never errors.
func rawBytesToJSON(b []byte) []byte {
	out, _ := json.Marshal(b)
	return out
}

// actorCallbackMetadataCtxKey carries the metadata of the actor callback
// being processed on the context handed to actor code.
type actorCallbackMetadataCtxKey struct{}

func withActorCallbackMetadata(ctx context.Context, md map[string]string) context.Context {
	if len(md) == 0 {
		return ctx
	}
	// Copy so the value carried on the context is independent of the protobuf
	// request's map and cannot be mutated through it.
	cp := make(map[string]string, len(md))
	for k, v := range md {
		cp[k] = v
	}
	return context.WithValue(ctx, actorCallbackMetadataCtxKey{}, cp)
}

// reentrancyIDFromContext returns the reentrancy id of the actor callback in
// flight on this context, if any.
func reentrancyIDFromContext(ctx context.Context) (string, bool) {
	md, ok := ctx.Value(actorCallbackMetadataCtxKey{}).(map[string]string)
	if !ok {
		return "", false
	}
	for k, v := range md {
		if strings.EqualFold(k, reentrancyIDMetadataKey) && v != "" {
			return v, true
		}
	}
	return "", false
}
