# Dapr Workflow Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Workflow

<!-- STEP
name: Run Workflow
output_match_mode: substring
expected_stdout_lines:
  - 'TestWorkflow registered'
  - 'TestActivity registered'
  - 'FailActivity registered'
  - 'Worker initialized'
  - 'runner started'
  - 'workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - 'workflow paused'
  - 'workflow resumed'
  - 'stage: 1'
  - 'workflow event raised'
  - 'stage: 2'
  - 'fail activity executions: 3'
  - 'workflow status: COMPLETED'
  - 'workflow purged'
  - 'stage: 2'
  - 'workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - 'workflow status: RUNNING'
  - 'workflow terminated'
  - 'workflow purged'
  - 'workflow worker successfully shutdown'

background: true
sleep: 60
timeout_seconds: 60
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
  - 'TestWorkflow registered'
  - 'TestActivity registered'
  - 'Worker initialized'
  - 'runner started'
  - 'workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - 'workflow paused'
  - 'workflow resumed'
  - 'stage: 1'
  - 'workflow event raised'
  - 'stage: 2'
  - 'workflow status: COMPLETED'
  - 'workflow purged'
  - 'stage: 2'
  - 'workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - 'workflow terminated'
  - 'workflow purged'
  - 'workflow client test'
  - '[wfclient] started workflow with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '[wfclient] workflow status: RUNNING'
  - '[wfclient] stage: 1'
  - '[wfclient] event raised'
  - '[wfclient] stage: 2'
  - '[wfclient] workflow terminated'
  - '[wfclient] workflow purged'
  - 'workflow worker successfully shutdown'
```
