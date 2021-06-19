package http

import (
	"net/http"

	"github.com/dapr/go-sdk/service/common"
)

// NewService creates new Service
func NewService(address string) common.Service {
	return newServer(address, nil)
}

// NewServiceWithMux creates new Service with existing http mux
func NewServiceWithMux(address string, mux *http.ServeMux) common.Service {
	return newServer(address, mux)
}

func newServer(address string, mux *http.ServeMux) *Server {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return &Server{
		address:            address,
		mux:                mux,
		topicSubscriptions: make([]*common.Subscription, 0),
	}
}

// Server is the HTTP server wrapping mux many Dapr helpers
type Server struct {
	address            string
	mux                *http.ServeMux
	topicSubscriptions []*common.Subscription
}

// Start starts the HTTP handler. Blocks while serving
func (s *Server) Start() error {
	s.registerSubscribeHandler()
	server := http.Server{
		Addr:    s.address,
		Handler: s.mux,
	}
	return server.ListenAndServe()
}

// Stop stops previously started HTTP service
func (s *Server) Stop() error {
	// TODO: implement service stop
	return nil
}

func setOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
	w.Header().Set("Allow", "POST,OPTIONS")
}

func optionsHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			setOptions(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	}
}
