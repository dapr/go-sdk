package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dapr/go-sdk/service"
)

// AddServiceInvocationHandler appends provided service invocation handler with its name to the service
func (s *ServiceImp) AddServiceInvocationHandler(name string, fn func(ctx context.Context, in *service.InvocationEvent) (out *service.InvocationEvent, err error)) error {
	if name == "" {
		return fmt.Errorf("service name required")
	}

	route := fmt.Sprintf("/%s", name)
	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var e *service.InvocationEvent

			// check for post with no data
			if r.ContentLength > 0 {
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if content != nil {
					e = &service.InvocationEvent{
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
