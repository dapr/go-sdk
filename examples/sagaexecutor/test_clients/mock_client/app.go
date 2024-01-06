package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	dapr "github.com/dapr/go-sdk/client"
	service "github.com/stevef1uk/sagaexecutor/service"
)

const myTopic = "test-client"

var client dapr.Client
var s service.Server

func callback(w http.ResponseWriter, r *http.Request) {
	var params service.Start_stop
	fmt.Printf("Callback invoked!\n")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewDecoder(r.Body).Decode(&params)

	// Here do what is necessary to recover this transaction)
	fmt.Printf("transaction callback invoked %v\n\n", params)
	json.NewEncoder(w).Encode("ok")
}

func main() {
	var err error

	/*pp_id       string    `json:"app_id"`
	  Service      string    `json:"service"`
	  Token        string    `json:"token"`
	  callback_service string    `json:"callback_service"`
	  Params       string    `json:"params"`
	  Timeout      int       `json:"timeout"`
	  Event        bool      `json:"event"`
	  LogTime      time.Time `json:"logtime"`*/

	appPort := "6000"
	if value, ok := os.LookupEnv("APP_PORT"); ok {
		appPort = value
	}
	router := mux.NewRouter()

	log.Println("setting up handler")
	router.HandleFunc("/callback", callback).Methods("POST", "OPTIONS")
	go http.ListenAndServe(":"+appPort, router)

	log.Println("About to send a couple of messages")

	client, err = dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	s = service.NewService(myTopic)
	defer s.CloseService()

	log.Println("Sleeping for a bit")
	time.Sleep(5 * time.Second)

	log.Println("Finished sleeping")

	err = s.SendStart(client, "mock-client", "test2", "abcdefg1234", "NotExpected", `{"ERROR":true}`, 10)
	if err != nil {
		log.Printf("First Publish error got %s", err)
	} else {
		log.Println("Successfully published first start message")
	}

	err = s.SendStop(client, "mock-client", "test2", "abcdefg1234")
	if err != nil {
		log.Printf("First Stop publish  error got %s", err)
	} else {
		log.Println("Successfully published first stop message")
	}

	// Check no records stored
	log.Println("Checking no records left")
	s.GetAllLogs(client, "mock-client", "test2")

	log.Println("Sending a Start without a Stop & waiting for the call-back")
	err = s.SendStart(client, "mock-client", "test2", "abcdefg1235", "callback", `{"Expected":TRUE}`, 30)
	if err != nil {
		log.Printf("Second Publish error got %s", err)
	} else {
		log.Println("Successfully published second start message")
	}
	// Check one record but no call back yet
	s.GetAllLogs(client, "mock-client", "test2")

	log.Println("Sleeping for a bit for the Poller to call us back ")
	time.Sleep(40 * time.Second)

	// Now lets test some load

	log.Println("Sending a group of starts & stops")
	for i := 0; i < 20; i++ {
		token := uuid.NewString()
		err = s.SendStart(client, "mock-client", "test2", token, "callback", `{"ERROR":Unexpected!}`, 20)
		if err != nil {
			log.Printf("First Publish error got %s", err)
		} else {
			log.Printf("Start %v - %s\n", i, token)
		}
		err = s.SendStop(client, "mock-client", "test2", token)
		if err != nil {
			log.Printf("First Stop publish  error got %s", err)
		} else {
			log.Printf("Stop %v - %s\n", i, token)
		}
	}
	log.Println("Finished sending starts & stops")
	log.Println("Sleeping for quite a bit to allow time to receive any callbacks")
	time.Sleep(60 * time.Second)

	client.Close()

}
