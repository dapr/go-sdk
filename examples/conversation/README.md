# Dapr Conversation Example

## Step

### Prepare

- Dapr installed

### Run Converse

<!-- STEP
name: Run Conversation Client
output_match_mode: substring
match_order: none
expected_stdout_lines:
  - '== APP == hi there'
background: false
sleep: 5
timeout_seconds: 60
-->

```bash
dapr run --app-id conversation-api\
         --log-level debug \
         --resources-path ./config/ \
         go run ./main.go
```

<!-- END_STEP -->


## Result
- Conversation output

```
== APP == hi there
✅  Exited App successfully
```
