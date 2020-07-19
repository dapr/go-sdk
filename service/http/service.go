package http

import (
	"context"
	"net/http"

	"github.com/dapr/go-sdk/service"
)

// NewService creates new Service
func NewService(address string) service.Service {
	return newService(address)
}

func newService(address string) *ServiceImp {
	return &ServiceImp{
		address:            address,
		mux:                http.NewServeMux(),
		topicSubscriptions: make([]*service.Subscription, 0),
	}
}

// ServiceImp is the HTTP server wrapping mux many Dapr helpers
type ServiceImp struct {
	address            string
	mux                *http.ServeMux
	server             http.Server
	topicSubscriptions []*service.Subscription
}

// Start starts the HTTP handler. Blocks while serving
func (s *ServiceImp) Start() error {
	s.registerSubscribeHandler()
	s.server = http.Server{
		Addr:    s.address,
		Handler: s.mux,
	}
	return s.server.ListenAndServe()
}

// Stop stops the previously started service
func (s *ServiceImp) Stop() error {
	return s.server.Shutdown(context.Background())
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
