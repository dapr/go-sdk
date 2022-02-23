# Hello World with Unix domain socket

This tutorial will demonstrate how to instrument your application with Dapr, and run it locally on your machine.
You will deploying advanced `order` applications with [Unix domain socket](https://en.wikipedia.org/wiki/Unix_domain_socket) based on [Hello World](https://github.com/dapr/go-sdk/tree/main/examples/hello-world).

There is a great performance imporvement With Unix domain socket, please notice that it does not support on Windows.

## Prerequisites
This quickstart requires you to have the following installed on your machine:
- [Docker](https://docs.docker.com/)
- [Go](https://golang.org/)

## Step 1 - Setup Dapr

Follow [instructions](https://docs.dapr.io/getting-started/install-dapr/) to download and install the Dapr CLI and initialize Dapr.

## Step 2 - Understand the code

The [order.go](./order.go) is a simple command line application, that implements four commands:
* `put` sends an order with configurable order ID.
* `get` return the current order number.
* `del` deletes the order.
* `seq` streams a sequence of orders with incrementing order IDs.

First, the app instantiates Dapr client:

```go
    client, err := dapr.NewClientWithSocket(socket)
    if err != nil {
        panic(err)
    }
    defer client.Close()
```

Then, depending on the command line argument, the app invokes corresponding method:

Persist the state:
```go
    err := client.SaveState(ctx, stateStoreName, "order", []byte(strconv.Itoa(orderID)), nil)
```
Retrieve the state:
```go
    item, err := client.GetState(ctx, stateStoreName, "order", nil)
```
Delete the state:
```go
    err := client.DeleteState(ctx, stateStoreName, "order", nil)
```

## Step 3 - Run the app with Dapr

1. Build the app

<!-- STEP
name: Build the app
-->

```bash
make
```

<!-- END_STEP -->

2. Run the app

There are two ways to launch Dapr applications. You can pass the app executable to the Dapr runtime:

<!-- STEP
name: Run and send order
background: true
sleep: 5
expected_stdout_lines:
  - '== APP == dapr client initializing for: /tmp/dapr-order-app-grpc.socket'
  - '== APP == Sending order ID 20'
  - '== APP == Successfully persisted state'
-->

```bash
dapr run --app-id order-app --log-level error --unix-domain-socket /tmp -- ./order put --id 20
```

<!-- END_STEP -->

<!-- STEP
name: Run and get order
background: true
sleep: 5
expected_stdout_lines:
  - '== APP == dapr client initializing for: /tmp/dapr-order-app-grpc.socket'
  - '== APP == Getting order'
  - '== APP == Order ID 20'
-->

```bash
dapr run --app-id order-app --log-level error --unix-domain-socket /tmp ./order get
```

<!-- END_STEP -->

Alternatively, you can start a standalone Dapr runtime, and call the app from another shell:

```bash
dapr run --app-id order-app --log-level error --unix-domain-socket /tmp
```


```bash
./order put --id 10

./order get
```

To terminate your services, simply stop the "dapr run" process, or use the Dapr CLI "stop" command:

```bash
dapr stop --app-id order-app
```


3. Run multiple apps

You can run more than one app in Dapr runtime. In this example you will call `order seq` which sends a sequence of orders.
Another instance of the `order` app will read the state.

```sh
dapr run --app-id order-app --log-level error --unix-domain-socket /tmp ./order seq
```

```sh
./order get
```
