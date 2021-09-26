package http

import (
	"context"
	"encoding/json"
	"fmt"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/runtime"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

const (
	// PubSubHandlerSuccessStatusCode is the successful ack code for pubsub event appcallback response.
	PubSubHandlerSuccessStatusCode int = http.StatusOK

	// PubSubHandlerRetryStatusCode is the error response code (nack) pubsub event appcallback response.
	PubSubHandlerRetryStatusCode int = http.StatusInternalServerError

	// PubSubHandlerDropStatusCode is the pubsub event appcallback response code indicating that Dapr should drop that message.
	PubSubHandlerDropStatusCode int = http.StatusSeeOther
)

func (s *Server) registerBaseHandler() {
	// register subscribe handler
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(s.topicSubscriptions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	s.mux.HandleFunc("/dapr/subscribe", f)

	// register health check handler
	fHealth := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/healthz", fHealth).Methods(http.MethodGet)

	// register actor config handler
	fRegister := func(w http.ResponseWriter, r *http.Request) {
		data, err := runtime.GetActorRuntimeInstance().GetJsonSerializedConfig()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err = w.Write(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/dapr/config", fRegister).Methods(http.MethodGet)

	// register actor method invoke handler
	fInvoke := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		methodName := varsMap["methodName"]
		reqData, _ := ioutil.ReadAll(r.Body)
		rspData, err := runtime.GetActorRuntimeInstance().InvokeActorMethod(actorType, actorID, methodName, reqData)
		if err == actorErr.ErrActorTypeNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != actorErr.Success {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rspData)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}/method/{methodName}", fInvoke).Methods(http.MethodPut)

	// register deactivate actor handler
	fDelete := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		err := runtime.GetActorRuntimeInstance().Deactive(actorType, actorID)
		if err == actorErr.ErrActorTypeNotFound || err == actorErr.ErrActorIDNotFound {
			w.WriteHeader(http.StatusNotFound)
		}
		if err != actorErr.Success {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}", fDelete).Methods(http.MethodDelete)

	// register actor reminder invoke handler
	fReminder := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		reminderName := varsMap["reminderName"]
		reqData, _ := ioutil.ReadAll(r.Body)
		err := runtime.GetActorRuntimeInstance().InvokeReminder(actorType, actorID, reminderName, reqData)
		if err == actorErr.ErrActorTypeNotFound {
			w.WriteHeader(http.StatusNotFound)
		}
		if err != actorErr.Success {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}/method/remind/{reminderName}", fReminder).Methods(http.MethodPut)

	// register actor timer invoke handler
	fTimer := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		timerName := varsMap["timerName"]
		reqData, _ := ioutil.ReadAll(r.Body)
		err := runtime.GetActorRuntimeInstance().InvokeTimer(actorType, actorID, timerName, reqData)
		if err == actorErr.ErrActorTypeNotFound {
			w.WriteHeader(http.StatusNotFound)
		}
		if err != actorErr.Success {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}/method/timer/{timerName}", fTimer).Methods(http.MethodPut)
}

// AddTopicEventHandler appends provided event handler with it's name to the service.
func (s *Server) AddTopicEventHandler(sub *common.Subscription, fn func(ctx context.Context, e *common.TopicEvent) (retry bool, err error)) error {
	if sub == nil {
		return errors.New("subscription required")
	}
	if sub.Topic == "" {
		return errors.New("topic name required")
	}
	if sub.PubsubName == "" {
		return errors.New("pub/sub name required")
	}
	if sub.Route == "" {
		return errors.New("handler route name")
	}
	if fn == nil {
		return fmt.Errorf("topic handler required")
	}

	if !strings.HasPrefix(sub.Route, "/") {
		sub.Route = fmt.Sprintf("/%s", sub.Route)
	}

	s.topicSubscriptions = append(s.topicSubscriptions, sub)

	s.mux.Handle(sub.Route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// check for post with no data
			if r.ContentLength == 0 {
				http.Error(w, "nil content", PubSubHandlerDropStatusCode)
				return
			}

			// deserialize the event
			var in common.TopicEvent
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, err.Error(), PubSubHandlerDropStatusCode)
				return
			}

			if in.Topic == "" {
				in.Topic = sub.Topic
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// execute user handler
			retry, err := fn(r.Context(), &in)
			if err == nil {
				writeStatus(w, common.SubscriptionResponseStatusSuccess)
				return
			}

			if retry {
				writeStatus(w, common.SubscriptionResponseStatusRetry)
				return
			}

			writeStatus(w, common.SubscriptionResponseStatusDrop)
		})))

	return nil
}

func writeStatus(w http.ResponseWriter, s string) {
	status := &common.SubscriptionResponse{Status: s}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), PubSubHandlerRetryStatusCode)
	}
}
