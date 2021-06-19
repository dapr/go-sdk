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
	if fn == nil {
		return fmt.Errorf("binding handler required")
	}

	if !strings.HasPrefix(route, "/") {
		route = fmt.Sprintf("/%s", route)
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var content []byte
			if r.ContentLength > 0 {
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				content = body
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
			in := &common.BindingEvent{
				Data:     content,
				Metadata: meta,
			}
			out, err := fn(r.Context(), in)
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
