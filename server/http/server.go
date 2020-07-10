package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/dapr/go-sdk/server/event"
)

// NewServer creates new Server
func NewServer(mux *http.ServeMux) (server *Server, err error) {
	if mux == nil {
		return nil, fmt.Errorf("nil http mux")
	}
	return &Server{
		mux:                mux,
		topicSubscriptions: make([]*subscription, 0),
	}, nil
}

// Server is the HTTP server wrapping gin with many Dapr helpers
type Server struct {
	mux                *http.ServeMux
	topicSubscriptions []*subscription
}

// AddTopicEventHandler adds provided handler to the local list subscriptions
func (s *Server) AddTopicEventHandler(topic, route string, handler func(ctx context.Context, e event.TopicEvent) error) error {
	if topic == "" {
		return errors.New("nil topic name")
	}
	if route == "" {
		return errors.New("nil route name")
	}

	sub := &subscription{
		Topic: topic,
		Route: route,
	}
	s.topicSubscriptions = append(s.topicSubscriptions, sub)

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			content, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var in event.TopicEvent
			if err := json.Unmarshal(content, &in); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if in.Topic == "" {
				in.Topic = topic
			}
			if err := handler(r.Context(), in); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		})))

	return nil
}

// HandleSubscriptions creates Dapr topic subscriptions
func (s *Server) HandleSubscriptions() error {
	s.mux.Handle("/dapr/subscribe", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(s.topicSubscriptions); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	))
	return nil
}

func optionsHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
			w.Header().Set("Allow", "POST,OPTIONS")
		} else {
			h.ServeHTTP(w, r)
		}
	}
}

type subscription struct {
	Topic string `json:"topic"`
	Route string `json:"route"`
}
