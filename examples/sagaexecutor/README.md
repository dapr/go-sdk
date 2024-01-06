# Dapr Example Project SagaExecutor

This project has been written to demonstrate the use of Dapr Building Blocks using the Go SDK. It delivers a Saga Executor Orchestrator Service. The high-level solution is shown below:

![Architecture Diagram](./img/Overview.png)

There are 3 logical components of this solution:
1. The Saga Service Code which provides an interface and is instantiated into the Calling Go Service code
2. The Saga Subscriber 
3. The Saga Poller

The Saga components are shown in Green and the Dapr building blockes in Blue.

The client Go service code is linked with the Saga Service code. 

This provides methods for the client service to publish Start & Stop messages to a queue managed by Dapr. This ensure that the latency to the consuming Go service is minimal.

The Saga Subscriber component reads these messages and stores them in a database, the Saga Log, using a Go native Postgres driver. Originally, I used the Dapr DataStore, but this is process specific so I switched to Postgres. Only Start messages are stored and these are deleted when a Stop message is received.

The Saga Poll queries the State store for Start messages that exist and for which the timeout period has elapsed. When found the client’s service call-back method recorded in the Start message is invoked. If successful the Start message in the state store is deleted to avoid a repeat of the call-back method.

These components use Darp capabilities to reduce the amount of code required:
```
gocloc .
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              11            183             62            651
YAML                            18              5              2            468
Markdown                         1             40              0            254
JSON                             3              0              0             50
Makefile                         4              1              0             25
BASH                             1              4              3             20
-------------------------------------------------------------------------------
TOTAL                           38            233             67           1468
-------------------------------------------------------------------------------
```

Dapr allows you to deploy the same microservices from your local machines to the cloud. Correspondingly, this project has instructions for deploying [locally](#Run-Locally) or in [Kubernetes](#Run-in-Kubernetes). 

## Run locally

To demonstrate this project running locally it can be run on your machine as follows:

```
From the sagaexecutor directory:
cd test_clients
dapr init --dev   # if dapr is not already running
./setuplocal.sh   # this creates a Postgres docker container and initialises it
dapr run -f dapr
```
This should start the two core Saga components, the Poller and the Subscriber plus the test_server example code. The output should look like this and will run until you terminate it.
```
<snip>
== APP - server-test == 2024/01/01 09:14:01 Sleeping for a bit
== APP - sagapoller == 2024/01/01 09:14:06 Hello I am called by cron!
== APP - sagapoller == 2024/01/01 09:14:06 Returned 0 records
== APP - sagapoller == 2024/01/01 09:14:11 Hello I am called by cron!
== APP - sagapoller == 2024/01/01 09:14:11 Returned 0 records
== APP - server-test == 2024/01/01 09:14:11 Successfully pubished a start message for later callback
== APP - sagasubscriber == eventHandler Ordering Key = 1
== APP - sagasubscriber == 2024/01/01 09:14:11 App_id = server-test
== APP - sagasubscriber == 2024/01/01 09:14:11 eventHandler: Message:{server-test test1 abcdefgh1235 callback {"fred":1} 10 true 0001-01-01 00:00:00 +0000 UTC}
== APP - sagasubscriber == 2024/01/01 09:14:11 Start Storing key = server-testtest1abcdefgh1235, data = {"app_id":server-test,"service":test1,"token":abcdefgh1235,"callback_service":callback,"params":{"fred":1},"event": true,"timeout":10,"logtime":2024-01-01 08:14:11.823429 +0000 UTC}
== APP - sagasubscriber == 2024/01/01 09:14:11 DB:Store Key = server-testtest1abcdefgh1235
== APP - server-test == 2024/01/01 09:14:11 Successfully pubished Second start message callback
== APP - server-test == 2024/01/01 09:14:11 Successfully pubished Second stop message to cancel the start
== APP - server-test == 2024/01/01 09:14:11 setting up handler
== APP - sagasubscriber == eventHandler Ordering Key = 2
== APP - sagasubscriber == 2024/01/01 09:14:11 App_id = server-test
== APP - sagasubscriber == 2024/01/01 09:14:11 eventHandler: Message:{server-test test1 abcdefgh1236 callback {"steve":1} 20 true 0001-01-01 00:00:00 +0000 UTC}
== APP - sagasubscriber == 2024/01/01 09:14:11 Start Storing key = server-testtest1abcdefgh1236, data = {"app_id":server-test,"service":test1,"token":abcdefgh1236,"callback_service":callback,"params":{"steve":1},"event": true,"timeout":20,"logtime":2024-01-01 08:14:11.867344 +0000 UTC}
== APP - sagasubscriber == 2024/01/01 09:14:11 DB:Store Key = server-testtest1abcdefgh1236
== APP - sagasubscriber == eventHandler Ordering Key = 3
== APP - sagasubscriber == Stop so will delete state with key: server-testtest1abcdefgh1236
== APP - sagasubscriber == 2024/01/01 09:14:11 App_id = server-test
== APP - sagasubscriber == 2024/01/01 09:14:11 eventHandler: Message:{server-test test1 abcdefgh1236   0 false 0001-01-01 00:00:00 +0000 UTC}
== APP - sagasubscriber == 2024/01/01 09:14:11 DB:Delete Key = server-testtest1abcdefgh1236
== APP - sagasubscriber == 2024/01/01 09:14:11 Deleted Log with key server-testtest1abcdefgh1236
== APP - sagapoller == 2024/01/01 09:14:16 Hello I am called by cron!
== APP - sagapoller == 2024/01/01 09:14:16 Returned 1 records
== APP - sagapoller == 2024/01/01 09:14:16 time_tmp = 2024-01-01 081411.823429 +0000 UTC}. time_tmp = 2024-01-01 081411
== APP - sagapoller == 2024/01/01 09:14:16 App_id = server-test
== APP - sagapoller == 2024/01/01 09:14:16 Token = abcdefgh1235, Elapsed value = 5.007655s, Compared value = 10
== APP - sagapoller == 2024/01/01 09:14:21 Hello I am called by cron!
== APP - sagapoller == 2024/01/01 09:14:21 Returned 1 records
== APP - sagapoller == 2024/01/01 09:14:21 time_tmp = 2024-01-01 081411.823429 +0000 UTC}. time_tmp = 2024-01-01 081411
== APP - sagapoller == 2024/01/01 09:14:21 App_id = server-test
== APP - sagapoller == 2024/01/01 09:14:21 Token = abcdefgh1235, Elapsed value = 10.010032s, Compared value = 10
== APP - sagapoller == 2024/01/01 09:14:21 Token abcdefgh1235, need to invoke callback callback
== APP - sagapoller == sendCallBack invoked with key server-testtest1abcdefgh1235, params = {server-test test1 abcdefgh1235 callback  10 false 2024-01-01 08:14:11 +0000 UTC}
== APP - sagapoller == sendCallBack App_ID = server-test, Method = callback
== APP - server-test == Yay callback invoked!
== APP - server-test == transaction callback invoked {server-test test1 abcdefgh1235 callback  10 false 2024-01-01 08:14:11 +0000 UTC}
== APP - server-test == 
== APP - sagapoller == 2024/01/01 09:14:21 DB:Delete Key = server-testtest1abcdefgh1235
== APP - sagapoller == Deleted Log with key: server-testtest1abcdefgh1235
== APP - sagapoller == 2024/01/01 09:14:26 Hello I am called by cron!
== APP - sagapoller == 2024/01/01 09:14:26 Returned 0 records
```

## Run in Kubernetes

To get started with running this proejct, there are some prerequisites:

### Prerequisites
1. A kubernetes cluster is required with dapr installed (dapr init -k)
2. Redis & Postgres must be installed on the cluster
3. Tilt is is used to deply the components (see: https://tilt.dev). However, manual deployment is possible. Please note that these files
   build images for my DockerHub repo, so you will need to change these images to suitable names and then update the           deployment/kubernetes.yaml files to reference the revised image names. Also, unless your cluster is running on arm64 hardware (Mx Macs or     RPIs etc, you will need to change the GOARCH value from arm64 to say amd64 for 64-bit X86 platforms.
4. Go installed (latest version)

I used a personal hosted k3s cluster running on RPi4s, with k3s depolyed, this seems fairly solid but a Cloud SaaS version is expected to be used for real use cases of this software.

To install Postgres on my home cluster I used the Postgres Operator, which configures a HA set-up by default. See:  https://github.com/zalando/postgres-operator/tree/master

As I am using an arm system I needed to change the image being deployed: Change: image: registry.opensource.zalan.do/acid/postgres-operator:v1.10.1 in manifests/postgres-operator.yaml to: ghcr.io/zalando/postgres-operator:v1.10.1

Then I created a DB for this project, which I called hasura - on mac/Linux):
```
  export POSTGRES=$(kubectl get secret postgres.acid-minimal-cluster.credentials.postgresql.acid.zalan.do -n postgres -o 'jsonpath={.data.password}' | base64 -d)
  kubectl port-forward acid-minimal-cluster-0 -n postgres 5432:5432
  psql --host localhost --username postgres
  create database hasura with owner postgres;
  \c hasura;
  create table sagastate ( key text PRIMARY KEY, value text );
  GRANT ALL PRIVILEGES ON DATABASE hasura to postgres;
  GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public to postgres;
```
The postgres password is required to create a kubernetes secret as the deploymnet manifests expect this e.g
```
kubectl create secret generic postgres-url --from-literal="postgres-url=postgresql://postgres:$POSTGRES@acid-minimal-cluster.postgres.svc.cluster.local:5432/hasura"
```
To install Redis I used this Helm script: 
```
helm install my-release oci://registry-1.docker.io/bitnamicharts/redis
export REDIS_PASSWORD=$(kubectl get secret --namespace default my-release-redis -o jsonpath="{.data.redis-password}" | base64 -d)
kubectl create secret generic redis --from-literal="redis-password=$REDIS_PASSWORD"
```
The structure of the projects is:
```
components
cmd 
    poller
    subscriber
database
service
utility
test_clients
    mock_server
    mock_client
```

Sadly, there is a need to find the IP Address of the Master Redis Pod (my-release-redis-master-0) and update the pubsub.yaml file in Components with this.

```
kubectl get pod my-release-redis-master-0  --template '{{.status.podIP}}'
```

Also in the pubsub.yaml file it is necessary for kubernetest to delete these two lines so that publishing error's don't occur
```
auth:
  secretStore: local-secret-store
```

Before running the core Subscriber & Postgres componnets the config files in components need to be applied to the cluster e.g
```
kubectl create -f components/.
```
(the following files need to be used: : cron.yaml, observability.yaml, statestore.yaml & pubsub.yaml)

First deploy & run the Subscribers & Poller components (tilt up and tilt down to undeploy)

Then the test clients can be run (mock_server, mock_client, mock_client2) to demonstrate (or see) if it is working (again tilt up)

If the mock_client is run the output should look like this:

```
apr client initializing for: 127.0.0.1:50001
2023/12/19 14:43:15 setting up handler
2023/12/19 14:43:15 About to send a couple of messages
2023/12/19 14:43:15 Sleeping for a bit
2023/12/19 14:43:20 Finished sleeping
2023/12/19 14:43:20 Successfully published first start message
2023/12/19 14:43:20 Successfully published first stop message
2023/12/19 14:43:20 Checking no records left
2023/12/19 14:43:20 Returned 0 records
2023/12/19 14:43:20 Sending a Start without a Stop & waiting for the call-back
2023/12/19 14:43:20 Successfully published second start message
2023/12/19 14:43:20 Returned 0 records
2023/12/19 14:43:20 Sleeping for a bit for the Poller to call us back
Yay callback invoked!
transaction callback invoked {mock-client test2 abcdefg1235 callback {"Param1":France} 30 false 2023-12-19 14:43:20 +0000 UTC}
2023/12/19 14:44:00 Sending a group of starts & stops
2023/12/19 14:44:01 Finished sending starts & stops
2023/12/19 14:44:01 Sleeping for quite a bit to allow time to receive any callbacks
```
I removed use of the Dapr Statestore and used Postgres directly having created my own table for Saga log entries as shown above.
The Subscriber & Poller components can't access the same Dapr State entries other than using Postgres. 

I also tested this with the GCP Pub/Sub and the updated pubsub.yaml for GCP is as below:
```
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: sagatxs
spec:
  type: pubsub.gcp.pubsub
  version: v1
  metadata:
  - name: topic
    value: "saxalogs"
  - name: subscription
    value: "subscription1"
  - name: type
    value: service_account
  - name: projectId
    value: <YOur GCP Project ID> 
  - name: identityProjectId
  - name: privateKeyId
    value: <Service Account Provate Key Id>
  - name: clientEmail
    value: <id>-compute@developer.gserviceaccount.com
  - name: clientId
    value: <Your Client Id> 
  - name: authUri
    value: https://accounts.google.com/o/oauth2/auth
  - name: tokenUri
    value: https://oauth2.googleapis.com/token
  - name: authProviderX509CertUrl
    value: https://www.googleapis.com/oauth2/v1/certs
  - name: clientX509CertUrl
    value: https://www.googleapis.com/robot/v1/metadata/x509/<PROJECT_NAME>.iam.gserviceaccount.com #replace PROJECT_NAME
  - name: privateKey
    value: "-----BEGIN PRIVATE KEY-----  <Insert Your Key Here> -----END PRIVATE KEY-----"
  - name: disableEntityManagement
    value: "false"
  - name: enableMessageOrdering
    value: "true"
  - name: orderingKey
    value: "OrderingKey"
  - name: maxReconnectionAttempts # Optional
    value: 30
  - name: connectionRecoveryInSec # Optional
    value: 2
  - name: deadLetterTopic # Optional
    value: myapp_dlq
  - name: maxDeliveryAttempts # Optional
    value: 5
```


It is possible to run the Subscriber & Poller is a seperate namespace, say saga, by deploying the component yaml files to it and deploying 
these components to it (tilt has a --namespace=saga flag). Then the consuming service needs to have the appropriae namespace added to the app_id parameter e.g.:
```
err = s.SendStart(client, "server-test.default", "test1", "abcdefgh1235", "callback", `{"fred":1}`, 20)
```
To support one Subscriber per client service the dynamic subscription capabilities of Dapr have been used.
The client service  must now pass a unique topic name when instantiating the service e.g.
```
s = service.NewService(myTopic)
```
Then there are two yaml config files required. One is the kubernetes deploymnet file for the Subscriber. This is duplicated with the name changed to be unique plus the PORT number made unique. The other one creates the actual Pub/Sub topic subscripton e.g.
```
apiVersion: dapr.io/v2alpha1
kind: Subscription
metadata:
  name: sub0
spec:
  topic: test-service
  routes:
    default: /receivemessage
  pubsubname: sagatxs
scopes:
- sagasubscriber
```

The relevant items need to align to the names in the other yaml files for the auto-wiring to work.

Of course one done't need to have separate Subscribers per service client, it is possible to configure the Subscription to point to whatever Subscriber is required to be run.

## Usage Scenarios

Assume a service is having to call two other services as part of a logical transaction:

### Happy Path
In this case everything is ok so the Stops logged by Service A will cancel out the Starts by the Saga Subscriber. There will be nothing to recover. The Saga Scheduler will ensure that the Saga Log is empty after Service A has successfully completed both calls.

![Happy Path](./img/HappyPath.png)


### Unhappy Path
In this case something has gone wrong with Service C. Assuming retries have happened and Service C is still not responding, the the Saga Poller will detect that there is an unmatched Start 2 message and after the configured timeout it will call the error call-back handler passed in the Start 2 message. This message can contain json data in addition to a GUUID token based by Service A that will enable Service A to take the appropriate error recovery. This will be service specific, but could involve reversing the change made by invoking Service B again.
At the end of the recovery processing the Saga Log will be empty. The Start 2 message will remain in the Saga Log until he error handler method in Service A has been invoked.

![Unhappy Path](./img/UnhappyPath.png)
