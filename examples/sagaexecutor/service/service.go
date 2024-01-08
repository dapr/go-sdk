package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/dapr/components-contrib/bindings/postgres"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/stevef1uk/sagaexecutor/database"
	"github.com/stevef1uk/sagaexecutor/encodedecode"
	"github.com/stevef1uk/sagaexecutor/utility"
)

const (
	PubsubComponentName     = "sagatxs"
	stateStoreComponentName = "sagalogs"
)

type Start_stop = utility.Start_stop

type service struct { // Needed don't delete
}

var the_db *postgres.Postgres
var message_count int = 1
var pubsub_topic string

// @Summary Creates an instace of the Saga service. The DATABASE_URL environmnet variable must be set to the postgres URL.
// @ID NewService
// @Produce nil
// @Param topic The name of the topic to use to send to the corresponding Subscriber
func NewService(topic string) Server {
	the_db, _ = database.OpenDBConnection(os.Getenv("DATABASE_URL"))
	pubsub_topic = topic
	return &service{}
}

// @Summary Closes the postgres database connection(s) and the dapr client instance.
// @ID CloseService
// @Produce nil
func (service) CloseService() {
	database.CloseDBConnection(context.Background(), the_db)
}

func getNextMessageOrder() string {
	message_index := strconv.Itoa(message_count)
	message_count++
	return message_index
}

func postMessage(client dapr.Client, app_id string, s utility.Start_stop) error {
	s_bytes, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("postMessage() failed to marshall start_stop struct %v, %s", s, err)
	}

	//encode := utility.EncodeData(s_bytes)
	m := &utility.OrderedMessage{OrderingField: getNextMessageOrder(), Data: s_bytes}

	err = client.PublishEvent(context.Background(), PubsubComponentName, pubsub_topic,
		&pubsub.Message{
			Data:        m.Data,
			OrderingKey: m.OrderingField,
		})
	if err != nil {
		return fmt.Errorf("postMessage() failed to publish start_stop struct %q", err)
	}

	return nil
}

// @Summary Creates a Saga Log Start entry based on the passed parameters. Internally, the primary key for this record will be a concatination of the the first three parameters
// @ID SendStart
// @Produce error
// @Param client The dapr client name
// @Param app_id A unique apllication ID for this client of the Saga service
// @Param service A unique name for the dapr service.
// @Param token A unique string for this client & service e.g. a GUID to identify this specific Saga Log entry
// @Param callback_service The name of the call back handler for the defined dapr service to invoke after any tmeout occurs
// @Param params A json string that will be passed back to the callback_service by the Poller component. This is expected to hold state information that can be used to roll-back the processing.
// @Param timeout The timeout value in seconds for the Poller to use when it reads the Saga Log stored for this call to determine expiry and a need to invoke the callback_service handler
func (service) SendStart(client dapr.Client, app_id string, service string, token string, callback_service string, params string, timeout int) error {
	// Base64 encode params as they should be a json string
	params = encodedecode.EncodeData([]byte(params))
	s1 := utility.Start_stop{App_id: app_id, Service: service, Token: token, Callback_service: callback_service, Params: params, Timeout: timeout, Event: utility.Start, LogTime: time.Now()}
	return postMessage(client, app_id, s1)
}

// @Summary Deletes a Saga Log Start entry based on the passed parameters.
// @ID SendStop
// @Produce error
// @Param client The dapr client name
// @Param app_id A unique apllication ID for this client of the Saga service
// @Param service A unique name for the dapr service.
// @Param token The value of the token passed on the preceeding call to SendStart to entry the Saga Log entry to delete
func (service) SendStop(client dapr.Client, app_id string, service string, token string) error {
	s1 := utility.Start_stop{App_id: app_id, Service: service, Callback_service: "", Token: token, Params: "", Timeout: 0, Event: utility.Stop}
	return postMessage(client, app_id, s1)
}

func (service) GetAllLogs(client dapr.Client, app_id string, service string) {

	var log_entry utility.Start_stop

	ret, err := database.GetStateRecords(context.Background(), the_db)
	if err != nil {
		log.Printf("Error reading state records %s", err)
		return
	}

	log.Printf("Returned %d records\n", len(ret))

	for i := 0; i < len(ret); i++ {
		res_entry := ret[i]
		log_entry = utility.ProcessRecord(res_entry, false)

		elapsed := time.Since(log_entry.LogTime)
		allowed_time := log_entry.Timeout

		log.Printf("Token = %s, Elapsed value = %v, Compared value = %v\n", log_entry.Token, elapsed, allowed_time)

		if time.Duration.Seconds(elapsed) > float64(allowed_time) {
			log.Printf("Token %s, need to invoke callback %s\n", log_entry.Token, log_entry.Callback_service)
			log_entry.Params = encodedecode.DecodeData(log_entry.Params)
			sendCallback(client, res_entry.Key, log_entry)
		}
	}
}

func sendCallback(client dapr.Client, key string, params utility.Start_stop) {

	data, _ := json.Marshal(params)
	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        data,
	}

	fmt.Printf("sendCallBack invoked with key %s, params = %v\n", key, params)
	fmt.Printf("sendCallBack App_ID = %s, Method = %s\n", params.App_id, params.Callback_service)

	_, err := client.InvokeMethodWithContent(context.Background(), params.App_id, params.Callback_service, "post", content)
	if err == nil {
		// Delivered so lets delete the Start record from the Store

		err = database.Delete(context.Background(), the_db, key)
		if err == nil {
			fmt.Println("Deleted Log with key:", key)
		}
	}
}

func (service) DeleteStateEntry(key string) error {
	return database.Delete(context.Background(), the_db, key)
}

func (service) StoreStateEntry(key string, value []byte) error {
	return database.StoreState(context.Background(), the_db, key, value)
}
