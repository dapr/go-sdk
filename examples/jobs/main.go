package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/types/known/anypb"

	daprc "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/examples/jobs/api"
	"github.com/dapr/go-sdk/service/common"
	daprs "github.com/dapr/go-sdk/service/grpc"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

// expectedJobRuns is the number of triggers this example waits for before it
// inspects and deletes the job.
const expectedJobRuns = 3

// jobRuns is signalled once per received trigger, so the example can wait on
// the triggers themselves instead of on wall-clock time.
var jobRuns = make(chan struct{}, expectedJobRuns)

func main() {
	server, err := daprs.NewService(":50070")
	if err != nil {
		logger.Fatalf("failed to start the server: %v", err)
	}

	if err = server.AddJobEventHandler("prod-db-backup", prodDBBackupHandler); err != nil {
		logger.Fatalf("failed to register job event handler: %v", err)
	}

	logger.Println("starting server")
	go func() {
		if err = server.Start(); err != nil {
			logger.Fatalf("failed to start server: %v", err)
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

	job := daprc.NewJob("prod-db-backup",
		daprc.WithJobSchedule("@every 1s"),
		daprc.WithJobRepeats(10),
		daprc.WithJobData(&anypb.Any{
			Value: jobData,
		}),
		daprc.WithJobConstantFailurePolicy(),
		daprc.WithJobConstantFailurePolicyMaxRetries(4),
		// The retry interval has to stay well below the lifetime of this
		// example, otherwise a single failed trigger can never be retried
		// before the program exits.
		daprc.WithJobConstantFailurePolicyInterval(time.Second),
	)

	// create the client
	client, err := daprc.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	err = client.ScheduleJobAlpha1(ctx, job)
	if err != nil {
		panic(err)
	}

	fmt.Println("schedulejob - success")

	// Wait for the job to actually fire. Sleeping for a fixed duration instead
	// makes the run depend on how long the sidecar took to start, and on
	// whether every trigger succeeded on its first attempt.
	if err = waitForJobRuns(expectedJobRuns, 20*time.Second); err != nil {
		panic(err)
	}

	resp, err := client.GetJobAlpha1(ctx, "prod-db-backup")
	if err != nil {
		panic(err)
	}
	fmt.Printf("getjob - resp: Name: %s, Schedule: %s, Repeats: %d, DueTime: %s, TTL: %s, Data: %v\n", resp.Name, *resp.Schedule, *resp.Repeats, *resp.DueTime, *resp.TTL, resp.Data) // parse

	err = client.DeleteJobAlpha1(ctx, "prod-db-backup")
	if err != nil {
		fmt.Printf("job deletion error: %v\n", err)
	} else {
		fmt.Println("deletejob - success")
	}

	if err = server.Stop(); err != nil {
		logger.Fatalf("failed to stop server: %v\n", err)
	}
}

// waitForJobRuns blocks until the handler has reported count triggers, or
// until timeout elapses.
func waitForJobRuns(count int, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for i := 0; i < count; i++ {
		select {
		case <-jobRuns:
		case <-deadline.C:
			return fmt.Errorf("timed out after %s waiting for %d job triggers, got %d", timeout, count, i)
		}
	}
	return nil
}

var jobCount atomic.Int64

func prodDBBackupHandler(ctx context.Context, job *common.JobEvent) error {
	var jobPayload api.DBBackup
	if err := json.Unmarshal(job.Data, &jobPayload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	fmt.Printf("job %d received:\n type: %v \n payload: %v\n", jobCount.Add(1)-1, job.JobType, jobPayload)

	// Never block the handler: the sidecar treats a slow response as a failure.
	select {
	case jobRuns <- struct{}{}:
	default:
	}
	return nil
}
