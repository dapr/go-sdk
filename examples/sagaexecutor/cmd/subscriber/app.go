// Listen to a topic and store the messages in the Dapr StateStore
package main

import (
	"context"
	"fmt"
	"net/http"

	//"fmt"
	"log"
	"os"
	"strconv"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	common "github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/stevef1uk/sagaexecutor/database"
	service "github.com/stevef1uk/sagaexecutor/service"
	utility "github.com/stevef1uk/sagaexecutor/utility"
)

const stateStoreComponentName = "sagalogs"

type dataElement struct {
	Data    string             `json:"data"`
	LogData utility.Start_stop `json:"logdata"`
}

var sub = &common.Subscription{
	PubsubName: service.PubsubComponentName,
	Topic:      "Dummy-Not-Used",
	Route:      "/receivemessage",
}

var sub_client dapr.Client

var the_service service.Server

func main() {
	var err error
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "7005"
	}

	the_service = service.NewService("") // Subscriber doesn't send messages to a topic just read them
	defer the_service.CloseService()

	sub_client, err = dapr.NewClient()
	if err != nil {
		panic(err)
	}

	// Create the new server on appPort and add a topic listener
	s := daprd.NewService(":" + appPort)
	err = s.AddTopicEventHandler(sub, eventHandler)
	if err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	//log.Printf("Starting the server using port %s'n", appPort)
	// Start the server
	err = s.Start()
	if err != nil && err != http.ErrServerClosed {
		sub_client.Close()
		log.Fatalf("error listenning: %v", err)
	}
	sub_client.Close()
}

func storeMessage(client dapr.Client, m *utility.Start_stop) error {
	var err error

	//log.Printf("storeMessage m = %v\n", m)

	key := m.App_id + m.Service + m.Token

	// Only store Starts
	if m.Event == utility.Start {
		t := time.Now().UTC()
		s1 := t.String()

		log_m := `{"app_id":` + m.App_id + ","
		log_m += `"service":` + m.Service + ","
		log_m += `"token":` + m.Token + ","
		log_m += `"callback_service":` + m.Callback_service + ","
		log_m += `"params":` + m.Params + ","
		log_m += `"event": true` + ","
		log_m += `"timeout":` + strconv.Itoa(m.Timeout) + ","
		log_m += `"logtime":` + s1 + "}"

		log.Printf("Start Storing key = %s, data = %s\n", key, log_m)

		// Save state into the state store
		//err = client.SaveState(context.Background(), stateStoreComponentName, key, []byte(log_m), nil)
		err = the_service.StoreStateEntry(key, []byte(log_m))
		if err != nil {
			log.Fatal(err)
		}
	} else { // Stop means we delete the corresponding Start entry
		// Delete state from the state store
		fmt.Printf("Stop so will delete state with key: %s\n", key)
		/*err = client.DeleteState(context.Background(), stateStoreComponentName, key, nil)
		if err != nil {
			log.Fatal(err)
		}*/
		err = the_service.DeleteStateEntry(key) // Yes I really want to delete the Start record now!
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Deleted Log with key %s\n", key)
	}

	//log.Printf("exit storeMessage\n")
	return err
}

func eventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {

	//fmt.Println("eventHandler received:", e.Data)
	//fmt.Printf("type of e.Data: %s\n", reflect.TypeOf(e.Data))

	var m map[string]interface{} = e.Data.(map[string]interface{})

	fmt.Printf("eventHandler Ordering Key = %s\n", m["OrderingKey"].(string))

	tmp := &database.StateRecord{Key: "", Value: m["Data"].(string)}
	message := utility.ProcessRecord(*tmp, true)
	message.LogTime, _ = time.Parse(time.RFC3339Nano, time.Now().String())

	log.Printf("eventHandler: Message:%v\n", message)

	err = storeMessage(sub_client, &message)
	if err != nil {
		log.Fatalf("Unable to store message %s", err)
	}

	return false, err
}
