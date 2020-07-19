package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

// AddServiceInvocationHandler appends provided service invocation handler with its route to the service
func (s *ServiceImp) AddServiceInvocationHandler(route string, fn func(ctx context.Context, in *InvocationEvent) (out *InvocationEvent, err error)) error {
	if route == "" {
		return fmt.Errorf("service route required")
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
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
			if o != nil && o.Data != nil {
				if _, err := w.Write(o.Data); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if o.ContentType != "" {
					w.Header().Set("Content-type", o.ContentType)
				}
			}
		})))

	return nil
}
