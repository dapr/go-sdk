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
  - 'work item listener started'
  - '== APP == RetryN  1'
  - '== APP == RetryN  2'
  - '== APP == RetryN  3'
  - '== APP == RetryN  4'
  - '== APP == RetryN  1'
  - '== APP == RetryN  2'
  - '== APP == RetryN  3'
  - '== APP == RetryN  4'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'

background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id workflow-taskexecutionid \
         --dapr-grpc-port 50001 \
         --log-level debug \
         --resources-path ./config \
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

```
  - '== APP == Workflow(s) and activities registered.'
  - 'work item listener started'
  - '== APP == RetryN  1'
  - '== APP == RetryN  2'
  - '== APP == RetryN  3'
  - '== APP == RetryN  4'
  - '== APP == RetryN  1'
  - '== APP == RetryN  2'
  - '== APP == RetryN  3'
  - '== APP == RetryN  4'
  - '== APP == workflow status: COMPLETED'
  - '== APP == workflow terminated'
  - '== APP == workflow purged'
```

