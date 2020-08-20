# Dapr HTTP Service SDK for Go

Start by importing Dapr go `service/http` package:

```go
daprd "github.com/dapr/go-sdk/service/http"
```

## Event Handling 

To handle events from specific topic, first create a Dapr service, add topic event handler, and start the service:

```go
s := daprd.NewService(":8080")

sub := &common.Subscription{
	PubsubName: "messages",
	Topic: "topic1",
	Route: "/events",
}
err := s.AddTopicEventHandler(sub, eventHandler)
if err != nil {
	log.Fatalf("error adding topic subscription: %v", err)
}

if err = s.Start(); err != nil && err != http.ErrServerClosed {
	log.Fatalf("error listening: %v", err)
}

func eventHandler(ctx context.Context, e *common.TopicEvent) error {
	log.Printf("event - PubsubName:%s, Topic:%s, ID:%s, Data: %v", e.PubsubName, e.Topic, e.ID, e.Data)
	return nil
}
```

## Service Invocation Handler 

To handle service invocations, create and start the Dapr service as in the above example. In this case though add the handler for service invocation: 

```go
s := daprd.NewService(":8080")

if err := s.AddServiceInvocationHandler("/echo", echoHandler); err != nil {
	log.Fatalf("error adding invocation handler: %v", err)
}

if err := s.Start(); err != nil {
    log.Fatalf("server error: %v", err)
}

func echoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("invocation parameter required")
		return
	}
	log.Printf(
		"echo - ContentType:%s, Verb:%s, QueryString:%s, %+v",
		in.ContentType, in.Verb, in.QueryString, string(in.Data),
	)
	out = &common.Content{
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
s := daprd.NewService(":8080")

if err := s.AddBindingInvocationHandler("/run", runHandler); err != nil {
	log.Fatalf("error adding binding handler: %v", err)
}

func runHandler(ctx context.Context, in *common.BindingEvent) (out []byte, err error) {
	log.Printf("binding - Data:%v, Meta:%v", in.Data, in.Metadata)
	return nil, nil
}
```

## Templates 

To accelerate your Dapr app development in go even further you can use one of the GitHub templates integrating the HTTP Dapr callback package:

* [dapr-http-event-subscriber-template](https://github.com/dapr/dapr-http-event-subscriber-template)
* [dapr-http-cron-handler-template](https://github.com/dapr/dapr-http-cron-handler-template)


## Contributing to Dapr go client 

See the [Contribution Guide](../../CONTRIBUTING.md) to get started with building and developing.
