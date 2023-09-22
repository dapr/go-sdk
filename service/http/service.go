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

package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	"github.com/dapr/go-sdk/actor/runtime"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// Options for the server.
type ServiceOptions struct {
	// Existing HTTP Mux
	Mux *chi.Mux
	// Protocol to use
	// Defaults to "http"
	Protocol common.ServiceProtocol
	// TLS certificate, if using HTTPS
	TLSCert string
	// TLS key, if using HTTPS
	TLSKey string
	// TLS configuration, if using HTTPS
	// This is an alternative to specifying "TLSCert" and "TLSKey"
	TLSConfig *tls.Config
}

func (o ServiceOptions) GetTLSConfig() (*tls.Config, error) {
	// If there's a TLSConfig in the options, return that
	if o.TLSConfig != nil {
		return o.TLSConfig, nil
	}

	// Create a new tls.Config
	conf := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load a TLS certificate and key from PEM on disk
	if o.TLSCert != "" && o.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(o.TLSCert, o.TLSKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate and key: %w", err)
		}
		conf.Certificates = []tls.Certificate{cert}
	} else {
		// Generate a self-signed TLS certificate
		cert, err := common.GenerateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed TLS certificate: %w", err)
		}
		conf.Certificates = []tls.Certificate{cert}
	}

	return conf, nil
}

// NewService creates new Service.
func NewService(address string) common.Service {
	// Cannot error with these options
	svc, _ := NewServiceWithOptions(address, ServiceOptions{})
	return svc
}

// NewServiceWithMux creates new Service with existing http mux.
func NewServiceWithMux(address string, mux *chi.Mux) common.Service {
	// Cannot error with these options
	svc, _ := NewServiceWithOptions(address, ServiceOptions{
		Mux: mux,
	})
	return svc
}

// NewServiceWithOptions creates a new Service with the given options.
func NewServiceWithOptions(address string, opts ServiceOptions) (common.Service, error) {
	if opts.Mux == nil {
		opts.Mux = chi.NewRouter()
	}
	if opts.Protocol == "" {
		// If there's no user-defined protocol, try reading from the APP_PROTOCOL env var
		opts.Protocol = common.ServiceProtocol(os.Getenv(common.AppProtocolEnvVar))
	}

	srv := &Server{
		address:        address,
		mux:            opts.Mux,
		topicRegistrar: make(internal.TopicRegistrar),
		authToken:      os.Getenv(common.AppAPITokenEnvVar),
	}

	switch strings.ToLower(string(opts.Protocol)) {
	case "http", "":
		srv.httpServer = &http.Server{
			Addr:              address,
			Handler:           opts.Mux,
			ReadHeaderTimeout: 30 * time.Second,
		}
	case "https":
		tlsConfig, err := opts.GetTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS configuration: %w", err)
		}
		srv.httpServer = &http.Server{
			Addr:              address,
			Handler:           opts.Mux,
			ReadHeaderTimeout: 30 * time.Second,
			TLSConfig:         tlsConfig,
		}
		srv.useTLS = true
	case "h2c":
		h2s := &http2.Server{}
		srv.httpServer = &http.Server{
			Addr:              address,
			Handler:           h2c.NewHandler(opts.Mux, h2s),
			ReadHeaderTimeout: 30 * time.Second,
		}
	default:
		return nil, fmt.Errorf("invalid protocol: %v", opts.Protocol)
	}

	return srv, nil
}

// Server is the HTTP server wrapping mux many Dapr helpers.
type Server struct {
	address        string
	mux            *chi.Mux
	httpServer     *http.Server
	topicRegistrar internal.TopicRegistrar
	authToken      string
	useTLS         bool
}

// Deprecated: Use RegisterActorImplFactoryContext instead.
func (s *Server) RegisterActorImplFactory(f actor.Factory, opts ...config.Option) {
	runtime.GetActorRuntimeInstance().RegisterActorFactory(f, opts...)
}

func (s *Server) RegisterActorImplFactoryContext(f actor.FactoryContext, opts ...config.Option) {
	runtime.GetActorRuntimeInstanceContext().RegisterActorFactory(f, opts...)
}

// Start starts the HTTP handler. Blocks while serving.
func (s *Server) Start() error {
	s.registerBaseHandler()

	if s.useTLS {
		// Certs and keys are already included in the server
		return s.httpServer.ListenAndServeTLS("", "")
	}
	return s.httpServer.ListenAndServe()
}

// Stop stops previously started HTTP service with a five second timeout.
func (s *Server) Stop() error {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctxShutDown)
}

func (s *Server) GracefulStop() error {
	return s.Stop()
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
