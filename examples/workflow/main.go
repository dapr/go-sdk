package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dapr/go-sdk/workflow"
)

func main() {
	wr, err := workflow.NewRuntime("localhost", "50001")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Runtime initialized")

	if err := wr.RegisterWorkflow(TestWorkflow); err != nil {
		log.Fatal(err)
	}

	fmt.Println("TestWorkflow registered")

	if err := wr.RegisterActivity(TestActivityStep1); err != nil {
		log.Fatal(err)
	}

	fmt.Println("TestActivityStep1 registered")

	if err := wr.RegisterActivity(TestActivityStep2); err != nil {
		log.Fatal(err)
	}

	fmt.Println("TestActivityStep2 registered")

	var wg sync.WaitGroup

	// start workflow runner
	fmt.Println("runner 1")
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := wr.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second * 5)

	// start workflow
	body, err := WorkflowHttp("start", "hi", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	if err != nil {
		fmt.Printf("failed to start workflow: %v\n", err.Error())
	}
	body, err = WorkflowHttp("pause", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	if err != nil {
		fmt.Printf("failed to pause workflow: %v\n", err.Error())
	}

	body, err = WorkflowHttp("get", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	if err != nil {
		fmt.Printf("failed to get workflow: %v\n", err.Error())
	}
	fmt.Printf("resp: %v\n", body)

	//// pause workflow
	//body, err = WorkflowHttp("pause", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to pause workflow: %v\n", err.Error())
	//}
	//body, err = WorkflowHttp("get", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to get workflow: %v\n", err.Error())
	//}
	//fmt.Println(body)
	//
	//// resume workflow
	//body, err = WorkflowHttp("resume", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to resume workflow: %v\n", err.Error())
	//}
	//
	//// raise event on workflow
	//body, err = WorkflowHttp("raiseEvent", "hi", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to raiseEvent on workflow: %v\n", err.Error())
	//}
	//
	//// purge workflow
	//// attempt to get the workflow
	//body, err = WorkflowHttp("purge", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to purge workflow: %v\n", err.Error())
	//}
	//
	//body, err = WorkflowHttp("get", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to get workflow: %v\n", err.Error())
	//}
	//fmt.Println(body)
	//
	//// start a new workflow for testing termination
	//// terminate and attempt get
	//body, err = WorkflowHttp("start", "hi", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to start workflow: %v\n", err.Error())
	//}
	//
	//body, err = WorkflowHttp("terminate", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to terminate workflow: %v\n", err.Error())
	//}
	//
	//body, err = WorkflowHttp("get", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	//if err != nil {
	//	fmt.Printf("failed to get workflow: %v\n", err.Error())
	//}

	// purge workflow
	_, err = WorkflowHttp("terminate", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	if err != nil {
		fmt.Printf("failed to terminate %v\n", err.Error())
	}
	body, err = WorkflowHttp("purge", "", "a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9")
	if err != nil {
		fmt.Printf("failed to purge workflow: %v\n", err.Error())
	}

	fmt.Printf("", body)

	time.Sleep(time.Second * 5)

	wg.Done()

	wg.Wait()
}

func WorkflowHttp(endpoint string, input string, id string) (body string, err error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	wfComponent := "dapr"
	wfName := "TestWorkflow"

	urlBase := "http://localhost:3500/v1.0-beta1/workflows"
	var url, method string
	switch endpoint {
	case "start":
		url = fmt.Sprintf("%s/%s/%s/start?instanceID=%s", urlBase, wfComponent, wfName, id)
		method = "POST"
	case "terminate":
		url = fmt.Sprintf("%s/%s/%s/terminate", urlBase, wfComponent, id)
		method = "POST"
	case "raiseEvent":
		url = fmt.Sprintf("%s/%s/%s/raiseEvent/TestEvent", urlBase, wfComponent, id)
		method = "POST"
	case "pause":
		url = fmt.Sprintf("%s/%s/%s/pause", urlBase, wfComponent, id)
		method = "POST"
	case "resume":
		url = fmt.Sprintf("%s/%s/%s/resume", urlBase, wfComponent, id)
		method = "POST"
	case "purge":
		url = fmt.Sprintf("%s/%s/%s/purge", urlBase, wfComponent, id)
		method = "POST"
	case "get":
		url = fmt.Sprintf("%s/%s/%s", urlBase, wfComponent, id)
		method = "GET"
	}

	var req *http.Request

	if endpoint == "start" || endpoint == "raiseEvent" {
		jsonBody := []byte(fmt.Sprintf("%q", input))
		bodyBytes := bytes.NewReader(jsonBody)

		fmt.Printf("Request body: %v\n", jsonBody)

		req, err = http.NewRequest(method, url, bodyBytes)
		if err != nil {
			return "", err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return "", err
		}
	}

	fmt.Println(url)

	req.Header.Set("dapr-app-id", "workflow-sequential")

	fmt.Printf("Request (%s) created\n", endpoint)

	// Invoking a service
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	fmt.Println("Request invoked with a response")

	// Reading response body
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("Status for (%s) request: %s\n", endpoint, resp.Status)

	return string(b), nil
}

func TestWorkflow(ctx *workflow.Context) (any, error) {
	var input string
	err := ctx.GetInput(&input)
	log.Printf("wf input %v\n", input)
	if err != nil {
		log.Printf("debug workflow err: %v\n", err)
		return nil, err
	}

	var output string
	err = ctx.CallActivity(TestActivityStep1).Await(&output)
	if err != nil {
		log.Printf(err.Error())
		return nil, err // TODO: populate error further
	}
	err = ctx.CallActivity(TestActivityStep2).Await(&output)
	if err != nil {
		log.Println(err.Error())
	}

	log.Printf("wf output: %v\n", output)
	log.Printf("name: %s, instanceid: %s, time: %v, replaying: %v\n", ctx.Name(), ctx.InstanceID(), ctx.CurrentUTCDateTime(), ctx.IsReplaying())
	// log.Printf("name: %s, in)
	return "test", nil // TODO: complete return
}

func TestActivityStep1(ctx workflow.ActivityContext) (any, error) {
	var input string
	// input may be empty
	err := ctx.GetInput(&input)
	if err != nil {
		// continue
	}

	log.Println("activity step 1 triggered")
	log.Printf("activity step  input: %v\n", input)

	return "step1", nil
}

func TestActivityStep2(ctx workflow.ActivityContext) (any, error) {
	var input string
	// input may be empty
	err := ctx.GetInput(&input)
	if err != nil {
		// continue
	}
	log.Println("activity step 2 triggered")
	log.Printf("activity step 2 input: %v\n", input)

	return "step2", nil
}
