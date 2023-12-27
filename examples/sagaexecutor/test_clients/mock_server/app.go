package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"net/http"
	"os"

	"github.com/gorilla/mux"

	dapr "github.com/dapr/go-sdk/client"
	service "github.com/stevef1uk/sagaexecutor/service"
)

const myTopic = "test-service"

var client dapr.Client
var s service.Server

func callback(w http.ResponseWriter, r *http.Request) {
	var params service.Start_stop
	fmt.Printf("Yay callback invoked!\n")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewDecoder(r.Body).Decode(&params)

	// Here do what is necessary to recover this transaction)
	fmt.Printf("transaction callback invoked %v\n\n", params)
	json.NewEncoder(w).Encode("ok")
}

func main() {
	var err error

	appPort := "6000"
	if value, ok := os.LookupEnv("APP_PORT"); ok {
		appPort = value
	}

	client, err = dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	s = service.NewService(myTopic)
	defer s.CloseService()

	log.Println("Sleeping for a bit")
	time.Sleep(10 * time.Second)

	// Now ensure that the Poller will call us back
	err = s.SendStart(client, "server-test", "test1", "abcdefgh1235", "callback", `{"fred":1}`, 20)
	if err != nil {
		log.Printf("First Publish error got %s", err)
	} else {
		log.Println("Successfully pubished a start message for later callback")
	}

	// Send a pair of Start & Stops messages so these shoud not result in a call-back
	err = s.SendStart(client, "server-test", "test1", "abcdefgh1236", "callback", `{"steve":1}`, 20)
	if err != nil {
		log.Printf("Second Publish error got %s", err)
	} else {
		log.Println("Successfully pubished Second start message callback")
	}
	err = s.SendStop(client, "server-test", "test1", "abcdefgh1236")
	if err != nil {
		log.Printf("Second Publish Stop error got %s", err)
	} else {
		log.Println("Successfully pubished Second stop message to cancel the start")
	}

	router := mux.NewRouter()
	log.Println("setting up handler")
	router.HandleFunc("/callback", callback).Methods("POST", "OPTIONS")
	log.Fatal(http.ListenAndServe(":"+appPort, router))
}
