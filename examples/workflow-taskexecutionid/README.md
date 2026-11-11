# Dapr Parallel Workflow Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Workflow

<!-- STEP
name: Run Workflow
output_match_mode: substring
expected_stdout_lines:
  - 'Workflow(s) and activities registered.'
  - 'RetryN  1'
  - 'RetryN  2'
  - 'RetryN  3'
  - 'RetryN  4'
  - 'RetryN  1'
  - 'RetryN  2'
  - 'RetryN  3'
  - 'RetryN  4'
  - 'workflow status: COMPLETED'
  - 'workflow terminated'
  - 'workflow purged'

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
  - 'Workflow(s) and activities registered.'
  - 'RetryN  1'
  - 'RetryN  2'
  - 'RetryN  3'
  - 'RetryN  4'
  - 'RetryN  1'
  - 'RetryN  2'
  - 'RetryN  3'
  - 'RetryN  4'
  - 'workflow status: COMPLETED'
  - 'workflow terminated'
  - 'workflow purged'
```

