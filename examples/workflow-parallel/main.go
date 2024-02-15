package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dapr/go-sdk/workflow"
)

func main() {
	w, err := workflow.NewWorker()
	if err != nil {
		log.Fatalf("failed to initialise worker: %v", err)
	}

	if err := w.RegisterWorkflow(BatchProcessingWorkflow); err != nil {
		log.Fatalf("failed to register workflow: %v", err)
	}
	if err := w.RegisterActivity(GetWorkBatch); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}
	if err := w.RegisterActivity(ProcessWorkItem); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}
	if err := w.RegisterActivity(ProcessResults); err != nil {
		log.Fatalf("failed to register activity: %v", err)
	}
	fmt.Println("Workflow(s) and activities registered.")

	if err := w.Start(); err != nil {
		log.Fatalf("failed to start worker")
	}

	wfClient, err := workflow.NewClient()
	if err != nil {
		log.Fatalf("failed to initialise client: %v", err)
	}
	ctx := context.Background()
	id, err := wfClient.ScheduleNewWorkflow(ctx, "BatchProcessingWorkflow", workflow.WithInput(10))
	if err != nil {
		log.Fatalf("failed to schedule a new workflow: %v", err)
	}

	metadata, err := wfClient.WaitForWorkflowCompletion(ctx, id)
	if err != nil {
		log.Fatalf("failed to get workflow: %v", err)
	}
	fmt.Printf("workflow status: %s\n", metadata.RuntimeStatus.String())

	err = wfClient.TerminateWorkflow(ctx, id)
	if err != nil {
		log.Fatalf("failed to terminate workflow: %v", err)
	}
	fmt.Println("workflow terminated")

	err = wfClient.PurgeWorkflow(ctx, id)
	if err != nil {
		log.Fatalf("failed to purge workflow: %v", err)
	}
	fmt.Println("workflow purged")
}

func BatchProcessingWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var input int
	if err := ctx.GetInput(&input); err != nil {
		return 0, err
	}

	var workBatch []int
	if err := ctx.CallActivity(GetWorkBatch, workflow.ActivityInput(input)).Await(&workBatch); err != nil {
		return 0, err
	}

	parallelTasks := workflow.NewTaskSlice(len(workBatch))
	for i, workItem := range workBatch {
		parallelTasks[i] = ctx.CallActivity(ProcessWorkItem, workflow.ActivityInput(workItem))
	}

	var outputs int
	for _, task := range parallelTasks {
		var output int
		err := task.Await(&output)
		if err == nil {
			outputs += output
		} else {
			return 0, err
		}
	}

	if err := ctx.CallActivity(ProcessResults, workflow.ActivityInput(outputs)).Await(nil); err != nil {
		return 0, err
	}

	return 0, nil
}

func GetWorkBatch(ctx workflow.ActivityContext) (any, error) {
	var batchSize int
	if err := ctx.GetInput(&batchSize); err != nil {
		return 0, err
	}
	batch := make([]int, batchSize)
	for i := 0; i < batchSize; i++ {
		batch[i] = i
	}
	return batch, nil
}

func ProcessWorkItem(ctx workflow.ActivityContext) (any, error) {
	var workItem int
	if err := ctx.GetInput(&workItem); err != nil {
		return 0, err
	}
	fmt.Printf("Processing work item: %d\n", workItem)
	time.Sleep(time.Second * 5)
	result := workItem * 2
	fmt.Printf("Work item %d processed. Result: %d\n", workItem, result)
	return result, nil
}

func ProcessResults(ctx workflow.ActivityContext) (any, error) {
	var finalResult int
	if err := ctx.GetInput(&finalResult); err != nil {
		return 0, err
	}
	fmt.Printf("Final result: %d\n", finalResult)
	return finalResult, nil
}
