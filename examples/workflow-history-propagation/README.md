# Dapr Workflow History Propagation Example

This example demonstrates how workflows can propagate their execution history
to child workflows and activities, enabling downstream consumers to inspect
the full (or partial) execution context of their caller.

## Workflow Architecture

```
MerchantCheckout (workflow)
├── ValidateMerchant (activity, no propagation)
└── ProcessPayment (child workflow, PropagateLineage)
    ├── ValidateCard (activity, no propagation)
    ├── CheckSpendingLimits (activity, no propagation)
    ├── FraudDetection (child workflow, PropagateLineage)
    │     → sees 15 events: MerchantCheckout + ProcessPayment
    └── SettlePayment (activity, PropagateOwnHistory)
          → sees 12 events: ProcessPayment only
```

### Propagation Scope

| Mode | What it sends | Use case |
|------|--------------|----------|
| `PropagateLineage()` | Caller's own events + any ancestor events it received | Full chain-of-custody verification |
| `PropagateOwnHistory()` | Caller's own events only (no ancestor chain) | Trust boundary — downstream only sees the immediate caller |

### Key Demonstration

- **FraudDetection** receives 15 events via `PropagateLineage()` — it can verify
  that `ValidateMerchant` ran in the top-level/grandparent workflow (MerchantCheckout),
  plus `ValidateCard` and `CheckSpendingLimits` ran in ProcessPayment.

- **SettlePayment** receives 12 events via `PropagateOwnHistory()` — it only sees
  ProcessPayment's events. The MerchantCheckout ancestral history is excluded.

## Running the Example

Propagation works in either of two modes:

- **Standalone / `dapr run`** — quickest path. Propagation runs end-to-end with
  no mTLS or signing setup. Each propagating dispatch logs a warning so
  operators are aware the chunks are unsigned and can't be cryptographically
  verified by receivers.
- **Kubernetes with mTLS + WorkflowHistorySigning** — production-grade path.
  Chunks travel between sidecars over mTLS and the workflow's own history is
  signed. No warning logs are emitted.

### Option A: Standalone (`dapr run`)

```bash
dapr init                                              # if you haven't already
cd examples/workflow-history-propagation
go build -o payment-app .
dapr run --app-id payment-app --resources-path config -- ./payment-app
```

Note: build the binary and run it directly rather than `go run .` to ensure Ctrl+C
properly allows `dapr run` to exit.

You'll see lines like:

```
[FraudDetection] Received propagated history: 15 events (scope: LINEAGE)
... level=warning msg="propagating unsigned workflow history to ..."
```

The warnings are expected — they're telling you that without
`WorkflowHistorySigning` enabled, the chunks aren't signed.

### Option B: Kubernetes with mTLS + signing

This path adds Sentry-issued mTLS for sidecar-to-sidecar traffic and turns on
`WorkflowHistorySigning` so propagated chunks travel within a signed
trust boundary.

#### Prerequisites

- A running Kubernetes cluster with `kubectl` context set (eg `kind`).
- `docker` to build the example app image.
- `dapr` CLI on your PATH.

#### 1. Install Dapr into the cluster with mTLS on

Use the latest RC that includes workflow history propagation. mTLS is on by
default for Helm/`dapr init -k` installs.

```bash
dapr init -k --runtime-version=1.18.0     # 1.18.0+
```

Wait for the dapr-system pods to be Running:

```bash
kubectl get pods -n dapr-system
```

Expected: `dapr-sentry`, `dapr-placement-server-0`, `dapr-scheduler-server-0`,
`dapr-operator`, `dapr-sidecar-injector`.

#### 2. Build and load the example app image

From this example's directory:

```bash
cd examples/workflow-history-propagation

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o payment-app .

docker build -t payment-app:dev .
kind load docker-image payment-app:dev          # kind clusters only; otherwise push to your registry
```

#### 3. Deploy the example

```bash
kubectl apply -f k8s/
```

This applies four manifests:

- `redis-deploy.yaml` — Redis Deployment + Service (`payment-app-redis`)
- `wf-store.yaml` — Dapr `state.redis` Component pointing at `payment-app-redis:6379`
- `signing-config.yaml` — Configuration CRD enabling `WorkflowHistorySigning`
- `app-deploy.yaml` — `payment-app` Deployment with Dapr sidecar annotations

#### 4. Watch the output

```bash
kubectl logs -l app=payment-app -c payment-app -f
```

Expected substrings (all should appear):

```
WORKFLOW HISTORY PROPAGATION DEMO
[MerchantCheckout] Starting checkout for merchant
[ValidateMerchant] Validating merchant
[ProcessPayment] Starting payment
[FraudDetection] Received propagated history: 15 events (scope: LINEAGE)
[FraudDetection] Verification: ValidateMerchant=true, ValidateCard=true, CheckSpendingLimits=true
[FraudDetection] APPROVED
[SettlePayment] Settling
propagated history: 12 events, scope=OWN_HISTORY
[SettlePayment] SETTLED
COMPLETE
```

In this mode the daprd sidecar should NOT log
`propagating unsigned workflow history` — the chunks are signed because
`WorkflowHistorySigning` is on.

### Troubleshooting

- **FraudDetection reports 0 events** — the workflow code didn't request
  propagation. Confirm the parent calls `CallChildWorkflow` /
  `CallActivity` with `workflow.WithHistoryPropagation(...)`.
- **`propagating unsigned workflow history` warnings in standalone** —
  expected; switch to the Kubernetes path above (or enable Sentry +
  `WorkflowHistorySigning`) if you want signed chunks.
- **K8s: Sentry connection errors** — check `kubectl get pods -n dapr-system`
  and re-deploy the control plane.

### Cleanup

Tear down the example resources (`payment-app` Deployment, `payment-app-redis`
Deployment + Service, `wf-store` Component, `signing` Configuration):

```bash
kubectl delete -f k8s/
```

Confirm nothing is left over:

```bash
kubectl get deploy,svc,po,component,configuration -l app=payment-app
kubectl get deploy,svc,po -l app=payment-app-redis
```

Both should return `No resources found`.

To uninstall the Dapr control plane (matching `dapr init -k`):

```bash
dapr uninstall -k --all          # removes dapr-system namespace + CRDs
```

## Files

```
workflow-history-propagation/
├── README.md                # this file
├── main.go                  # workflow + activity definitions
├── config/
│   └── redis.yaml           # local self-hosted Dapr Component (dapr run)
├── Dockerfile               # packages the pre-built payment-app binary
└── k8s/
    ├── signing-config.yaml  # Configuration CRD enabling WorkflowHistorySigning
    ├── wf-store.yaml        # Dapr Component (state.redis, actorStateStore)
    ├── redis-deploy.yaml    # Redis Deployment + Service (payment-app-redis)
    └── app-deploy.yaml      # payment-app Deployment with dapr sidecar annotations
```
