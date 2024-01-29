# Dapr Workflow Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Workflow

<!-- STEP
name: Run Workflow
output_match_mode: substring
expected_stdout_lines:
  - '== APP == Worker initialized'
  - '== APP == TestWorkflow registered'
  - '== APP == TestActivity registered'
  - '== APP == runner started'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow paused'
  - '== APP == workflow resumed'
  - '== APP == stage: 1'
  - '== APP == workflow event raised'
  - '== APP == stage: 2'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow purged'
  - '== APP == stage: 2'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'
  - '== APP == workflow client test'
  - '== APP == [wfclient] started workflow with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == [wfclient] workflow status: RUNNING'
  - '== APP == [wfclient] stage: 1'
  - '== APP == [wfclient] event raised'
  - '== APP == [wfclient] stage: 2'
  - '== APP == [wfclient] workflow terminated'
  - '== APP == [wfclient] workflow purged'
  - '== APP == workflow worker successfully shutdown'

background: true
sleep: 60
-->

```bash
dapr run --app-id workflow \
         --dapr-grpc-port 50001 \
         --log-level debug \
         --resources-path ./config \
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

```
  - '== APP == Worker initialized'
  - '== APP == TestWorkflow registered'
  - '== APP == TestActivity registered'
  - '== APP == runner started'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow paused'
  - '== APP == workflow resumed'
  - '== APP == stage: 1'
  - '== APP == workflow event raised'
  - '== APP == stage: 2'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow purged'
  - '== APP == stage: 2'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'
  - '== APP == workflow client test'
  - '== APP == [wfclient] started workflow with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == [wfclient] workflow status: RUNNING'
  - '== APP == [wfclient] stage: 1'
  - '== APP == [wfclient] event raised'
  - '== APP == [wfclient] stage: 2'
  - '== APP == [wfclient] workflow terminated'
  - '== APP == [wfclient] workflow purged'
  - '== APP == workflow worker successfully shutdown'
```
