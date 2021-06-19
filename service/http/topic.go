package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/dapr/go-sdk/service/common"
)

const (
	// PubSubHandlerSuccessStatusCode is the successful ack code for pubsub event appcallback response
	PubSubHandlerSuccessStatusCode int = http.StatusOK

	// PubSubHandlerRetryStatusCode is the error response code (nack) pubsub event appcallback response
	PubSubHandlerRetryStatusCode int = http.StatusInternalServerError

	// PubSubHandlerDropStatusCode is the pubsub event appcallback response code indicating that Dapr should drop that message
	PubSubHandlerDropStatusCode int = http.StatusSeeOther
)

func (s *Server) registerSubscribeHandler() {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(s.topicSubscriptions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	s.mux.HandleFunc("/dapr/subscribe", f)
}

// AddTopicEventHandler appends provided event handler with it's name to the service
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)) error {
	if sub == nil {
		return errors.New("subscription required")
	}
	if sub.Topic == "" {
		return errors.New("topic name required")
	}
	if sub.PubsubName == "" {
		return errors.New("pub/sub name required")
	}
	if sub.Route == "" {
		return errors.New("handler route name")
	}
	if fn == nil {
		return fmt.Errorf("topic handler required")
	}

	if !strings.HasPrefix(sub.Route, "/") {
		sub.Route = fmt.Sprintf("/%s", sub.Route)
	}

	s.topicSubscriptions = append(s.topicSubscriptions, sub)

	s.mux.Handle(sub.Route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// check for post with no data
			if r.ContentLength == 0 {
				http.Error(w, "nil content", PubSubHandlerDropStatusCode)
				return
			}

			// deserialize the event
			var in common.TopicEvent
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), PubSubHandlerDropStatusCode)
				return
			}

			if in.Topic == "" {
				in.Topic = sub.Topic
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// execute user handler
			retry, err := fn(r.Context(), &in)
			if err == nil {
				writeStatus(w, common.SubscriptionResponseStatusSuccess)
				return
			}

			if retry {
				writeStatus(w, common.SubscriptionResponseStatusRetry)
				return
			}

			writeStatus(w, common.SubscriptionResponseStatusDrop)
		})))

	return nil
}

func writeStatus(w http.ResponseWriter, s string) {
	status := &common.SubscriptionResponse{Status: s}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), PubSubHandlerRetryStatusCode)
	}
}
