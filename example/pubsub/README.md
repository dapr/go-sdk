# Dapr PubSub Example with go-sdk

This folder contains two go file that uses this go-SDK to invoke the Dapr PubSub API.

## Helpful
![](https://i.loli.net/2020/08/23/5MBYgwqCZcXNUf2.jpg)

## Step

### Prepare
- Dapr installed
- Run a redis with host: 127.0.0.1:6379 password: 123456

### Run Subscriber Server
when we use Dapr PubSub to subscribe, we should have a http or gRPC server to receive the requests from Dapr.

Please change directory to pubsub/ and run the following command:
```bash
dapr run --app-id sub \ 
         --app-protocol http \ 
         --app-port 8080 \ 
         --port 3500 \ 
         --log-level debug \ 
         --components-path ./config \ 
         go run sub.go 
```

### Run Publisher
Publish is more simply than subscribe. Just Publish the data to target pubsub component with its' name.

After you start a server by above guide.
Please change directory to pubsub/ and run the following command:
```bash
dapr run --app-id pub \ 
         --log-level debug \ 
         --components-path ./config \ 
         go run pub.go 
```

## Result
You would see log that in terminal which run the server(subscriber) code.
```bash
== APP == 2020/08/23 13:21:58 event - PubsubName: messagebus, Topic: demo, ID: 11acaa82-23c4-4244-8969-7360dae52e5d, Data: ping
```