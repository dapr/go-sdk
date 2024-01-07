// Componnet to scan for timedout uncompleted transactions and  invoke call-backs
package main

import (
	"context"
	"log"

	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

func closeAll() {
	client.Close()
	the_service.CloseService()
}

func multiSignalHandler(signal os.Signal) {

	switch signal {
	case syscall.SIGHUP:
		logger.Println("Signal:", signal.String())
		closeAll()
		os.Exit(0)
	case syscall.SIGINT:
		closeAll()
		logger.Println("Signal:", signal.String())
		os.Exit(0)
	case syscall.SIGTERM:
		logger.Println("Signal:", signal.String())
		closeAll()
		os.Exit(0)
	case syscall.SIGQUIT:
		closeAll()
		logger.Println("Signal:", signal.String())
		os.Exit(0)
	default:
		logger.Println("Unhandled/unknown signal")
	}
}

func main() {
	// create a Dapr service
	s := daprd.NewService(address)

	client, err = dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// add some input binding handler
	if err := s.AddBindingInvocationHandler("sagapoller", sagaHandler); err != nil {
		logger.Fatalf("error adding binding handler: %v", err)
	}

	the_service = service.NewService("")
	defer the_service.CloseService()

	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM) //we can add more sycalls.SIGQUIT etc.
	exitchnl := make(chan int)

	go func() {
		for {
			s := <-sigchnl
			multiSignalHandler(s)
		}
	}()

	// start the service
	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("error starting service: %v", err)
	}

	exitcode := <-exitchnl
	os.Exit(exitcode)
}

func sagaHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	log.Println("sagaHandler I am called by cron!")
	the_service.GetAllLogs(client, "", "")
	return nil, nil
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
