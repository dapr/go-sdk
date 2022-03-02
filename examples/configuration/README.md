# Dapr Configuration Example

## Step

### Prepare

- Dapr installed

### Run Get Configuration

<!-- STEP
name: Run Configuration Client
output_match_mode: substring
expected_stdout_lines:
  - '== APP == get config = myConfigValue'
  - '== APP == get updated config key = mySubscribeKey1, value = mySubscribeValue1'
  - '== APP == get updated config key = mySubscribeKey2, value = mySubscribeValue1'
  - '== APP == get updated config key = mySubscribeKey3, value = mySubscribeValue1'
  - '== APP == get updated config key = mySubscribeKey1, value = mySubscribeValue2'
  - '== APP == get updated config key = mySubscribeKey2, value = mySubscribeValue2'
  - '== APP == get updated config key = mySubscribeKey3, value = mySubscribeValue2'
  - '== APP == get updated config key = mySubscribeKey1, value = mySubscribeValue3'
  - '== APP == get updated config key = mySubscribeKey2, value = mySubscribeValue3'
  - '== APP == get updated config key = mySubscribeKey3, value = mySubscribeValue3'
  - '== APP == dapr configuration subscribe finished.'
background: false
sleep: 40
-->

```bash
dapr run --app-id configuration-api\
         --app-protocol grpc \
         --app-port 5005 \
         --dapr-http-port 3006 \
         --log-level debug \
         --components-path ./config/ \
         go run ./main.go
```

<!-- END_STEP -->


## Result
- Configuration Client Logs

The subscription event order may out of order.

```
get config = myConfigValue
get updated config key = mySubscribeKey1, value = mySubscribeValue1 
get updated config key = mySubscribeKey2, value = mySubscribeValue1 
get updated config key = mySubscribeKey3, value = mySubscribeValue1 
get updated config key = mySubscribeKey1, value = mySubscribeValue2 
get updated config key = mySubscribeKey2, value = mySubscribeValue2 
get updated config key = mySubscribeKey3, value = mySubscribeValue2 
get updated config key = mySubscribeKey1, value = mySubscribeValue3 
get updated config key = mySubscribeKey2, value = mySubscribeValue3 
get updated config key = mySubscribeKey3, value = mySubscribeValue3 
dapr configuration subscribe finished.
✅  Exited App successfully

```
