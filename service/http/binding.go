package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dapr/go-sdk/service"
)

// AddBindingInvocationHandler appends provided binding invocation handler with its name to the service
func (s *ServiceImp) AddBindingInvocationHandler(name string, fn func(ctx context.Context, in *service.BindingEvent) (out []byte, err error)) error {
	if name == "" {
		return fmt.Errorf("binding name required")
	}

	route := fmt.Sprintf("/%s", name)
	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var e service.BindingEvent
			if r.ContentLength > 0 {
				// deserialize the event
				if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				e = service.BindingEvent{}
			}

			// execute handler
			out, err := fn(r.Context(), &e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if out == nil {
				out = []byte("{}")
			}

			w.Header().Add("Content-Type", "application/json")
			if _, err := w.Write(out); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})))

	return nil
}
