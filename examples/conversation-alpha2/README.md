# Dapr Conversation (Alpha1) Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Conversation Example

<!-- STEP
name: Run Conversation
output_match_mode: substring
expected_stdout_lines:
  - '== APP == conversation input: hello world'
  - '== APP == conversation output: hello world'

background: true
sleep: 120
timeout_seconds: 120
-->

```bash
dapr run --app-id conversation 
         --dapr-grpc-port 50001 
         --log-level debug 
         --resources-path ./config 
         -- go run ./main.go
```

<!-- END_STEP -->

## Result

```
  - '== APP == conversation output: hello world'
```
