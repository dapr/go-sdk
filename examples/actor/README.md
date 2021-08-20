# Dapr Actor Example with go-sdk

## Step

### Prepare

- Dapr installed

### Run Actor Server


```bash
dapr run --app-id actor-serving \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --components-path ./config \
         go run serving/main.go
```

<!-- END_STEP -->

### Run Actor Client

```bash
dapr run --app-id actor-client \
         --log-level debug \
         --components-path ./config \
         go run client/main.go
```

<!-- END_STEP -->

### Cleanup

```bash
dapr stop --app-id  actor-serving
```

<!-- END_STEP -->

## Result
- client side
```shell
== APP == dapr client initializing for: 127.0.0.1:55776
== APP == get user result =  &{abc 123}
== APP == get invoke result =  laurence
== APP == get post result =  laurence
== APP == get result =  get result
== APP == start timer
== APP == stop timer
== APP == start reminder
== APP == stop reminder
✅  Exited App successfully

```


- server side
```shell

== APP == call get user req =  &{abc 123}
== APP == get req =  laurence
== APP == get post request =  laurence
== APP == get req =  hello
== APP == get req =  hello
== APP == get req =  hello
== APP == get req =  hello
== APP == get req =  hello
== APP == receive reminder =  testReminderName  state =  "hello" duetime =  5s period =  5s
== APP == receive reminder =  testReminderName  state =  "hello" duetime =  5s period =  5s
== APP == receive reminder =  testReminderName  state =  "hello" duetime =  5s period =  5s
== APP == receive reminder =  testReminderName  state =  "hello" duetime =  5s period =  5s
== APP == receive reminder =  testReminderName  state =  "hello" duetime =  5s period =  5s
== APP == receive reminder =  testReminderName  state =  "hello" duetime =  5s period =  5s
```