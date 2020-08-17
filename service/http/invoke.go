package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/dapr/go-sdk/service/common"
)

// AddServiceInvocationHandler appends provided service invocation handler with its route to the service
func (s *Server) AddServiceInvocationHandler(route string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	if route == "" {
		return fmt.Errorf("service route required")
	}

	if !strings.HasPrefix(route, "/") {
		route = fmt.Sprintf("/%s", route)
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// capture http args
			e := &common.InvocationEvent{
				Verb:        r.Method,
				QueryString: valuesToMap(r.URL.Query()),
				ContentType: r.Header.Get("Content-type"),
			}

			// check for post with no data
			if r.ContentLength > 0 {
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				e.Data = content
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

func valuesToMap(in url.Values) map[string]string {
	out := map[string]string{}
	for k := range in {
		out[k] = in.Get(k)
	}
	return out
}
