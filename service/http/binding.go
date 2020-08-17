package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dapr/go-sdk/service/common"
)

// AddBindingInvocationHandler appends provided binding invocation handler with its route to the service
func (s *Server) AddBindingInvocationHandler(route string, fn func(ctx context.Context, in *common.BindingEvent) (out []byte, err error)) error {
	if route == "" {
		return fmt.Errorf("binding route required")
	}

	if !strings.HasPrefix(route, "/") {
		route = fmt.Sprintf("/%s", route)
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			content, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// assuming Dapr doesn't pass multiple values for key
			meta := map[string]string{}
			for k, values := range r.Header {
				// TODO: Need to figure out how to parse out only the headers set in the binding + Traceparent
				// if k == "raceparent" || strings.HasPrefix(k, "dapr") {
				for _, v := range values {
					meta[k] = v
				}
				// }
			}

			// execute handler
			out, err := fn(r.Context(), &common.BindingEvent{Data: content, Metadata: meta})
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
