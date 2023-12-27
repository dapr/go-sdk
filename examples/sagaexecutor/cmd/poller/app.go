// Componnet to scan for timedout uncompleted transactions and  invoke call-backs
package main

import (
	"context"
	"log"

	"net/http"
	"os"
	"strings"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	service "github.com/stevef1uk/sagaexecutor/service"
)

var (
	logger      = log.New(os.Stdout, "", 0)
	address     = getEnvVar("ADDRESS", ":8080")
	the_service service.Server
	client      dapr.Client
	err         error
)

func main() {
	// create a Dapr service
	s := daprd.NewService(address)

	client, err = dapr.NewClient()
	if err != nil {
		panic(err)
	}

	// add some input binding handler
	if err := s.AddBindingInvocationHandler("sagapoller", sagaHandler); err != nil {
		logger.Fatalf("error adding binding handler: %v", err)
	}

	the_service = service.NewService()
	defer the_service.CloseService()

	// start the service
	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("error starting service: %v", err)
	}
}

func sagaHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	//logger.Printf("Binding - Metadata:%v, Data:%v", in.Metadata, in.Data)
	log.Println("Hello I am called by cron!")

	// TODO: do something with the cloud event data
	the_service.GetAllLogs(client, "", "")

	return nil, nil
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
