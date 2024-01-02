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
  - '== APP == TestActivityStep1 registered'
  - '== APP == TestActivityStep2 registered'
  - '== APP == Status for (start) request: 202 Accepted'
  - 'Created new workflow instance with ID'
background: true
sleep: 30
-->

```bash
dapr run --app-id workflow-sequential \
         --app-protocol grpc \
         --dapr-grpc-port 50001 \
         --dapr-http-port 3500 \
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
    - '== APP == TestActivityStep1 registered'
    - '== APP == TestActivityStep2 registered'
    - '== APP == Status for (start) request: 202 Accepted'
    - 'Created new workflow instance with ID 'a7a4168d-3a1c-41da-8a4f-e7f6d9c718d9'
```