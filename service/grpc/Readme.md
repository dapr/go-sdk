# Dapr gRPC Service SDK for Go

Start by importing Dapr go `service/grpc` package:

```go
daprd "github.com/dapr/go-sdk/service/grpc"
```

## Event Handling 

To handle events from specific topic, first create a Dapr service, add topic event handler, and start the service:

```go
s, err := daprd.NewService(":50001")
if err != nil {
    log.Fatalf("failed to start the server: %v", err)
}

if err := s.AddTopicEventHandler("messages", "topic1", eventHandler); err != nil {
    log.Fatalf("error adding topic subscription: %v", err)
}

if err := s.Start(); err != nil {
    log.Fatalf("server error: %v", err)
}

func eventHandler(ctx context.Context, e *daprd.TopicEvent) error {
	log.Printf("event - Topic:%s, ID:%s, Data: %v", e.Topic, e.ID, e.Data)
	return nil
}
```

## Service Invocation Handler 

To handle service invocations, create and start the Dapr service as in the above example. In this case though add the handler for service invocation: 

```go
s, err := daprd.NewService(":50001")
if err != nil {
    log.Fatalf("failed to start the server: %v", err)
}

if err := s.AddServiceInvocationHandler("echo", echoHandler); err != nil {
    log.Fatalf("error adding invocation handler: %v", err)
}

if err := s.Start(); err != nil {
    log.Fatalf("server error: %v", err)
}

func echoHandler(ctx context.Context, in *daprd.InvocationEvent) (out *daprd.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}
	log.Printf(
		"echo - ContentType:%s, Verb:%s, QueryString:%s, %+v",
		in.ContentType, in.Verb, in.QueryString, string(in.Data),
	)
	out = &daprd.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}
```

## Binding Invocation Handler 

To handle binding invocations, create and start the Dapr service as in the above examples. In this case though add the handler for binding invocation: 

```go
s, err := daprd.NewService(":50001")
if err != nil {
    log.Fatalf("failed to start the server: %v", err)
}

if err := s.AddBindingInvocationHandler("run", runHandler); err != nil {
    log.Fatalf("error adding binding handler: %v", err)
}

if err := s.Start(); err != nil {
    log.Fatalf("server error: %v", err)
}

func runHandler(ctx context.Context, in *daprd.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%v, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
```


## Contributing to Dapr go client 

See the [Contribution Guide](../../CONTRIBUTING.md) to get started with building and developing.
