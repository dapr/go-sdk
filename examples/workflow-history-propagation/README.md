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

## Step

### Pre-req

- Dapr initialized and running locally (`dapr init`)

### Run Workflow

<!-- STEP
name: Run Workflow
output_match_mode: substring
expected_stdout_lines:
  - 'WORKFLOW HISTORY PROPAGATION DEMO'
  - '[MerchantCheckout] Starting checkout for merchant'
  - '[ValidateMerchant] Validating merchant'
  - '[ProcessPayment] Starting payment'
  - '[FraudDetection] Received propagated history: 15 events (scope: LINEAGE)'
  - '[FraudDetection] Verification: ValidateMerchant=true, ValidateCard=true, CheckSpendingLimits=true'
  - '[FraudDetection] APPROVED'
  - '[SettlePayment] Settling'
  - 'propagated history: 12 events, scope=OWN_HISTORY'
  - '[SettlePayment] SETTLED'
  - 'COMPLETE'

background: true
sleep: 30
timeout_seconds: 60
-->

```bash
dapr run --app-id payment-app \
         --dapr-grpc-port 50001 \
         --resources-path ./config \
         -- go run .
```

<!-- END_STEP -->
