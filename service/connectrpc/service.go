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

package connectrpc

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/dapr/dapr/pkg/proto/runtime/v1/runtimeconnect"
	"github.com/go-chi/chi/v5"

	"github.com/dapr/go-sdk/actor"
	"github.com/dapr/go-sdk/actor/config"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/internal"
)

// NewService creates new Service.
func NewService(address string) common.Service {
	return newService(address, nil)
}

// NewServiceWithMux creates new Service with existing http mux.
func NewServiceWithMux(address string, mux *chi.Mux) common.Service {
	return newService(address, mux)
}

func newService(address string, router *chi.Mux) *Server {
	if router == nil {
		router = chi.NewRouter()
	}

	s := &Server{
		invokeHandlers:   make(map[string]common.ServiceInvocationHandler),
		topicRegistrar:   make(internal.TopicRegistrar),
		bindingHandlers:  make(map[string]common.BindingInvocationHandler),
		jobEventHandlers: make(map[string]common.JobEventHandler),
		authToken:        os.Getenv(common.AppAPITokenEnvVar),
		httpServer: &http.Server{ //nolint:gosec
			Addr:    address,
			Handler: router,
		},
		mux: router,
	}

	path, handler := runtimeconnect.NewAppCallbackHandler(s)
	router.Handle(path+"*", handler)
	path, handler = runtimeconnect.NewAppCallbackAlphaHandler(s)
	router.Handle(path+"*", handler)
	path, handler = runtimeconnect.NewAppCallbackHealthCheckHandler(s)
	router.Handle(path+"*", handler)

	return s
}

// Server is the gRPC service implementation for Dapr.
type Server struct {
	runtimeconnect.UnimplementedAppCallbackHandler
	runtimeconnect.UnimplementedAppCallbackHealthCheckHandler
	invokeHandlers     map[string]common.ServiceInvocationHandler
	topicRegistrar     internal.TopicRegistrar
	bindingHandlers    map[string]common.BindingInvocationHandler
	jobEventHandlers   map[string]common.JobEventHandler
	healthCheckHandler common.HealthCheckHandler
	authToken          string
	started            uint32
	httpServer         *http.Server
	mux                *chi.Mux
}

// Deprecated: Use RegisterActorImplFactoryContext instead.
func (s *Server) RegisterActorImplFactory(f actor.Factory, opts ...config.Option) {
	panic("Actor is not supported by gRPC API")
}

func (s *Server) RegisterActorImplFactoryContext(f actor.FactoryContext, opts ...config.Option) {
	panic("Actor is not supported by gRPC API")
}

// Start registers the server and starts it.
func (s *Server) Start() error {
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New("a gRPC server can only be started once")
	}
	return s.httpServer.ListenAndServe()
}

// Stop stops the previously-started service.
func (s *Server) Stop() error {
	if atomic.LoadUint32(&s.started) == 0 {
		return nil
	}

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctxShutDown)
}

// GracefulStop stops the previously-started service gracefully.
func (s *Server) GracefulStop() error {
	return s.Stop()
}
