package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/dapr/go-sdk/service/common"
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
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn func(ctx context.Context, e *common.TopicEvent) error) error {
	if sub == nil {
		return errors.New("subscription required")
	}
	if sub.Topic == "" {
		return errors.New("topic name required")
	}
	if sub.Route == "" {
		return errors.New("handler route name")
	}

	s.topicSubscriptions = append(s.topicSubscriptions, sub)

	s.mux.Handle(sub.Route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// check for post with no data
			if r.ContentLength == 0 {
				http.Error(w, "nil content", http.StatusBadRequest)
				return
			}

			// deserialize the event
			var in common.TopicEvent
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if in.Topic == "" {
				in.Topic = sub.Topic
			}

			if err := fn(r.Context(), &in); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})))

	return nil
}
