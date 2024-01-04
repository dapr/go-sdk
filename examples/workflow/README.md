# Dapr Workflow Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Workflow

<!-- STEP
name: Run Workflow
output_match_mode: substring
expected_stdout_lines:
  - '== APP == Runtime initialized'
  - '== APP == TestWorkflow registered'
  - '== APP == TestActivity registered'
  - '== APP == runner 1'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow paused'
  - '== APP == workflow resumed'
  - '== APP == workflow event raised'
  - '== APP == stage: 2'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow purged'
  - '== APP == stage: 2'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'
background: true
sleep: 30
-->

```bash
dapr run --app-id workflow-sequential \
         --app-protocol grpc \
         --dapr-grpc-port 50001 \
         --placement-host-address localhost:50005 \
         --log-level debug \
         --resources-path ./config \
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

- workflow

```
  - '== APP == Runtime initialized'
  - '== APP == TestWorkflow registered'
  - '== APP == TestActivity registered'
  - '== APP == runner 1'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow paused'
  - '== APP == workflow resumed'
  - '== APP == workflow event raised'
  - '== APP == stage: 2'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow purged'
  - '== APP == stage: 2'
  - '== APP == workflow started with id: a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'
```