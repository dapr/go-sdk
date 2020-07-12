package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// Service is the HTTP service wrapper
type Service interface {
	// AddTopicEventHandler appends to the service a pub/sub event handler for specific topic name
	AddTopicEventHandler(topic, route string, handler func(ctx context.Context, e TopicEvent) error) error
	// AddInvocationHandler appends to the service a external invocation handler for specific method name
	AddInvocationHandler(route string, fn func(ctx context.Context, in *InvocationEvent) (out []byte, err error)) error
	// Start starts the HTTP handler. Blocks while serving
	Start(address string) error
}

// NewService creates new Service
func NewService() Service {
	return newService()
}

func newService() *ServiceImp {
	return &ServiceImp{
		Mux:                http.NewServeMux(),
		topicSubscriptions: make([]*Subscription, 0),
	}
}

// ServiceImp is the HTTP server wrapping mux many Dapr helpers
type ServiceImp struct {
	Mux                *http.ServeMux
	topicSubscriptions []*Subscription
}

// Start starts the HTTP handler. Blocks while serving
func (s *ServiceImp) Start(address string) error {
	if address == "" {
		return errors.New("nil address")
	}

	s.registerSubscribeHandler()

	server := http.Server{
		Addr:    address,
		Handler: s.Mux,
	}

	return server.ListenAndServe()
}

func (s *ServiceImp) registerSubscribeHandler() {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(s.topicSubscriptions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	s.Mux.HandleFunc("/dapr/subscribe", f)
}

// AddInvocationHandler adds provided handler to the local collection before server start
func (s *ServiceImp) AddInvocationHandler(route string, fn func(ctx context.Context, in *InvocationEvent) (out []byte, err error)) error {
	if route == "" {
		return errors.New("nil route name")
	}

	s.Mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var e *InvocationEvent

			// check for post with no data
			if r.ContentLength > 0 {
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if content != nil {
					e = &InvocationEvent{
						ContentType: r.Header.Get("Content-type"),
						Data:        content,
					}
				}
			}

			// execute handler
			o, err := fn(r.Context(), e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// write to response if handler returned data
			if o != nil {
				if _, err := w.Write(o); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		})))

	return nil
}

// AddTopicEventHandler adds provided handler to the local list subscriptions
func (s *ServiceImp) AddTopicEventHandler(topic, route string, handler func(ctx context.Context, e TopicEvent) error) error {
	if topic == "" {
		return errors.New("nil topic name")
	}
	if route == "" {
		return errors.New("nil route name")
	}

	sub := &Subscription{Topic: topic, Route: route}
	s.topicSubscriptions = append(s.topicSubscriptions, sub)

	s.Mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			content, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var in TopicEvent
			if err := json.Unmarshal(content, &in); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if in.Topic == "" {
				in.Topic = topic
			}

			if err := handler(r.Context(), in); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})))

	return nil
}

func optionsHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
			w.Header().Set("Allow", "POST,OPTIONS")
		} else {
			h.ServeHTTP(w, r)
		}
	}
}
