package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	daprc "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/dist-scheduler/api"
	"github.com/dapr/go-sdk/service/common"
	daprs "github.com/dapr/go-sdk/service/grpc"
)

func main() {
	server, err := daprs.NewService(":50070")
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}

	if err = server.AddJobEventHandler("prod-db-backup", prodDBBackupHandler); err != nil {
		log.Fatalf("failed to register job event handler: %v", err)
	}

	log.Println("starting server")
	go func() {
		if err = server.Start(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Brief intermission to allow for the server to initialize.
	time.Sleep(10 * time.Second)

	ctx := context.Background()

	jobData, err := json.Marshal(&api.DBBackup{
		Task: "db-backup",
		Metadata: api.Metadata{
			DBName:         "my-prod-db",
			BackupLocation: "/backup-dir",
		},
	},
	)
	if err != nil {
		panic(err)
	}

	job := daprc.Job{
		Name:     "prod-db-backup",
		Schedule: "@every 1s",
		Repeats:  10,
		Data: &anypb.Any{
			Value: jobData,
		},
	}

	// create the client
	client, err := daprc.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	err = client.ScheduleJobAlpha1(ctx, &job)
	if err != nil {
		panic(err)
	}

	fmt.Println("schedulejob - success")

	time.Sleep(3 * time.Second)

	resp, err := client.GetJobAlpha1(ctx, "prod-db-backup")
	if err != nil {
		panic(err)
	}
	fmt.Printf("getjob - resp: %v\n", resp) // parse

	err = client.DeleteJobAlpha1(ctx, "prod-db-backup")
	if err != nil {
		fmt.Printf("job deletion error: %v\n", err)
	} else {
		fmt.Println("deletejob - success")
	}

	if err = server.Stop(); err != nil {
		log.Fatalf("failed to stop server: %v\n", err)
	}
}

var jobCount = 0

func prodDBBackupHandler(ctx context.Context, job *common.JobEvent) error {
	var jobData common.Job
	if err := json.Unmarshal(job.Data, &jobData); err != nil {
		return fmt.Errorf("failed to unmarshal job: %v", err)
	}
	decodedPayload, err := base64.StdEncoding.DecodeString(jobData.Value)
	if err != nil {
		return fmt.Errorf("failed to decode job payload: %v", err)
	}
	var jobPayload api.DBBackup
	if err := json.Unmarshal(decodedPayload, &jobPayload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	fmt.Printf("job %d received:\n type: %v \n typeurl: %v\n value: %v\n extracted payload: %v\n", jobCount, job.JobType, jobData.TypeURL, jobData.Value, jobPayload)
	jobCount++
	return nil
}
