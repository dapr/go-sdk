# Dapr Configuration Example

## Step

### Prepare

- Dapr installed

### Run Get Configuration

<!-- STEP
name: Run Configuration Client
output_match_mode: substring
match_order: none
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
         --resources-path ./config/ \
         go run ./main.go
```

<!-- END_STEP -->


## Result
- Configuration Client Logs

The subscription event order may out of order.

```
got config key = mykey, with value = myConfigValue

got config key = mySubscribeKey1, with value = mySubscribeValue1 
got config key = mySubscribeKey2, with value = mySubscribeValue1 
got config key = mySubscribeKey3, with value = mySubscribeValue1 
got config key = mySubscribeKey1, with value = mySubscribeValue2 
got config key = mySubscribeKey2, with value = mySubscribeValue2 
got config key = mySubscribeKey3, with value = mySubscribeValue2 
got config key = mySubscribeKey1, with value = mySubscribeValue3 
got config key = mySubscribeKey2, with value = mySubscribeValue3 
got config key = mySubscribeKey3, with value = mySubscribeValue3 
got config key = mySubscribeKey1, with value = mySubscribeValue4 
got config key = mySubscribeKey2, with value = mySubscribeValue4 
got config key = mySubscribeKey3, with value = mySubscribeValue4 
got config key = mySubscribeKey1, with value = mySubscribeValue5 
got config key = mySubscribeKey2, with value = mySubscribeValue5 
got config key = mySubscribeKey3, with value = mySubscribeValue5 
dapr configuration subscribe finished.
dapr configuration unsubscribed
✅  Exited App successfully

```
