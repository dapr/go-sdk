This project has been created to demonstrate the use of Dapr Building Blocks There is a solution architecture picture and description in the README.pdf file.
I have performed basic testing and it seems to work as I expect, but anyone using this should test & assess for their own needs.

Since the original version I have now modified this code so that each client service can have its own Subscriber listening to a dedicated topic.

There is actually very little code required:
```
gocloc .
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              10            179             73            693
YAML                            13              9              1            384
Markdown                         1             29              0            146
Makefile                         4              1              0             25
-------------------------------------------------------------------------------
TOTAL                           28            218             74           1248
-------------------------------------------------------------------------------
```

To get started with running this proejct, there are some prerequisites:

1. A kubernetes cluster is required with dapr installed (dapr init -k)
2. Redis & Postgres must be installed on the cluster
3. Tilt is is used to deply the components (see: https://tilt.dev). However, manual deployment is possible.

I used a personal hosted k3s cluster running on RPi4s, with k3s depolyed, this seems fairly solid but a Cloud SaaS version is expected to be used for real use cases of this software.

To install Postgres on my home cluster I used the Postgres Operator, which configures a HA set-up by default. See:  https://github.com/zalando/postgres-operator/tree/master

As I am using an arm system I needed to change the image being deployed: Change: image: registry.opensource.zalan.do/acid/postgres-operator:v1.10.1 in manifests/postgres-operator.yaml to: ghcr.io/zalando/postgres-operator:v1.10.1

Then I created a DB for this project, which I called hasura - on mac/Linux):
```
  export POSTGRES=$(kubectl get secret postgres.acid-minimal-cluster.credentials.postgresql.acid.zalan.do -n postgres -o 'jsonpath={.data.password}' | base64 -d)
  kubectl port-forward acid-minimal-cluster-0 -n postgres 5432:5432
  psql --host localhost --username postgres
  create database hasura with owner postgres;
  create table sagastate ( key text PRIMARY KEY, value jsonb );
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

    








