package http

import (
	"context"
	"encoding/json"
	"fmt"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/runtime"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dapr/go-sdk/service/common"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	// PubSubHandlerSuccessStatusCode is the successful ack code for pubsub event appcallback response
	PubSubHandlerSuccessStatusCode int = http.StatusOK

	// PubSubHandlerRetryStatusCode is the error response code (nack) pubsub event appcallback response
	PubSubHandlerRetryStatusCode int = http.StatusInternalServerError

	// PubSubHandlerDropStatusCode is the pubsub event appcallback response code indicating that Dapr should drop that message
	PubSubHandlerDropStatusCode int = http.StatusSeeOther
)

func (s *Server) registerSubscribeHandler() {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(s.topicSubscriptions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	s.mux.HandleFunc("/dapr/subscribe", f)
}

func (s *Server) registerActorHandler() {
	fRegister := func(w http.ResponseWriter, r *http.Request) {
		w.Write(runtime.GetActorRuntime().GetSerializedConfig())
		fmt.Println("get dapr/config invoke: " + string(runtime.GetActorRuntime().GetSerializedConfig()))
		w.WriteHeader(200)
	}
	s.mux.HandleFunc("/dapr/config", fRegister)

	fInvoke := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		methodName := varsMap["methodName"]
		reqData, _ := ioutil.ReadAll(r.Body)
		rspData, err := runtime.GetActorRuntime().InvokeActorMethod(actorType, actorID, methodName, reqData)
		if err == actorErr.ErrorActorTypeNotFound || err == actorErr.ErrorActorIDNotFound {
			w.WriteHeader(404)
		}
		if err != actorErr.Success {
			w.WriteHeader(500)
		}
		w.WriteHeader(200)
		w.Write(rspData)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}/method/{methodName}", fInvoke).Methods("PUT")

	fDelete := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		err := runtime.GetActorRuntime().Deactive(actorType, actorID)
		if err == actorErr.ErrorActorTypeNotFound || err == actorErr.ErrorActorIDNotFound {
			w.WriteHeader(404)
		}
		if err != actorErr.Success {
			w.WriteHeader(500)
		}
		w.WriteHeader(200)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}", fDelete).Methods("DELETE")

	fReminder := func(w http.ResponseWriter, r *http.Request) {
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		reminderName := varsMap["reminderName"]
		reqData, _ := ioutil.ReadAll(r.Body)
		err := runtime.GetActorRuntime().InvokeReminder(actorType, actorID, reminderName, reqData)
		if err == actorErr.ErrorActorTypeNotFound || err == actorErr.ErrorActorIDNotFound {
			w.WriteHeader(404)
		}
		if err != actorErr.Success {
			w.WriteHeader(500)
		}
		w.WriteHeader(200)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}/method/remind/{reminderName}", fReminder).Methods("PUT")

	fTimer := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("daprd sdk actor invoke timer ")
		varsMap := mux.Vars(r)
		actorType := varsMap["actorType"]
		actorID := varsMap["actorId"]
		timerName := varsMap["timerName"]
		reqData, _ := ioutil.ReadAll(r.Body)
		err := runtime.GetActorRuntime().InvokeTimer(actorType, actorID, timerName, reqData)
		if err == actorErr.ErrorActorTypeNotFound || err == actorErr.ErrorActorIDNotFound {
			w.WriteHeader(404)
		}
		if err != actorErr.Success {
			w.WriteHeader(500)
		}
		w.WriteHeader(200)
	}
	s.mux.HandleFunc("/actors/{actorType}/{actorId}/method/timer/{timerName}", fTimer).Methods("PUT")

	fHealth := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
	s.mux.HandleFunc("/healthz", fHealth)
}

// AddTopicEventHandler appends provided event handler with it's name to the service
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
				fmt.Println(err.Error())
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
