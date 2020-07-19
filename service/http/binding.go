package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AddBindingInvocationHandler appends provided binding invocation handler with its route to the service
func (s *ServiceImp) AddBindingInvocationHandler(route string, fn func(ctx context.Context, in *BindingEvent) (out []byte, err error)) error {
	if route == "" {
		return fmt.Errorf("binding route required")
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var e BindingEvent
			if r.ContentLength > 0 {
				// deserialize the event
				if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				e = BindingEvent{}
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
