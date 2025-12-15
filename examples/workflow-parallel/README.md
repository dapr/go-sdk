# Dapr Parallel Workflow Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Workflow

<!-- STEP
name: Run Workflow
output_match_mode: substring
expected_stdout_lines:
  - '== APP == Workflow(s) and activities registered.'
  - '== APP == Processing work item: 9'
  - '== APP == Work item 9 processed. Result: 18'
  - '== APP == Final result: 90'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'

background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id workflow-parallel \
         --dapr-grpc-port 50001 \
         --log-level debug \
         --resources-path ./config \
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

```
  - '== APP == Workflow(s) and activities registered.'
  - '== APP == Processing work item: 9'
  - '== APP == Work item 9 processed. Result: 18'
  - '== APP == Final result: 90'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'
```

