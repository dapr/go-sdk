# Workflow compensation

This example books a flight, hotel, and car in order. By default the car booking
fails, so the workflow cancels the hotel and then the flight. Successful bookings
register their compensations; failed or unattempted bookings are not cancelled.

The workflow remains `FAILED` after compensation, preserving the original booking
error. Each compensation uses a retry policy with up to three attempts. If one
compensation still fails, the remaining compensations are attempted and the errors
are combined. Compensation is a business operation, not a database rollback.

## Run

Install the Dapr CLI and initialize a local Dapr environment with Redis, placement,
and scheduler available (`dapr init`). The Go toolchain must satisfy
[`../go.mod`](../go.mod). Run these commands from this directory.

The activities only simulate bookings and print their results. They do not contact
real booking services.

<!-- STEP
name: Compensate failed trip
output_match_mode: substring
expected_stdout_lines:
  - 'Booked trip-demo-flight'
  - 'Booked trip-demo-hotel'
  - 'Cancelled trip-demo-hotel'
  - 'Cancelled trip-demo-flight'
  - 'workflow status: FAILED'
timeout_seconds: 90
-->

```bash
dapr run --app-id workflow-compensation \
         --dapr-grpc-port 50001 \
         --resources-path ./config \
         -- go run .
```

<!-- END_STEP -->

To exercise the successful path, pass `-fail-car=false`:

<!-- STEP
name: Book successful trip
output_match_mode: substring
expected_stdout_lines:
  - 'Booked trip-demo-flight'
  - 'Booked trip-demo-hotel'
  - 'Booked trip-demo-car'
  - 'workflow status: COMPLETED'
timeout_seconds: 90
-->

```bash
dapr run --app-id workflow-compensation \
         --dapr-grpc-port 50001 \
         --resources-path ./config \
         -- go run . -fail-car=false
```

<!-- END_STEP -->

Each run schedules a new workflow instance. The worker is bounded by a 60-second
context, which is cancelled when the program exits.

## Adapting the pattern

`BookTripWorkflow` connects the SDK to `bookTrip`, which keeps the ordered booking
and reverse-order compensation logic in one place. Workflow replay reconstructs
the local compensation list from activity results in history. External side
effects belong in activities, never in the workflow function.

Replace the simulated activities with your service calls. Use stable idempotency
keys, store reservations durably, and make cancellation idempotent. A real booking
can succeed remotely even if its activity fails before recording the result; such
uncertain outcomes require reconciliation or a cancellation that tolerates a
missing reservation. After compensation retries are exhausted, retain the failure
for operator investigation rather than reporting a successful rollback.

## Unit tests

```bash
go test -race .
```

The tests exercise booking success, failures at each step, reverse compensation
order, reservation propagation, and continuing after a compensation failure. They
do not require a sidecar. The README steps in the examples CI matrix additionally
exercise the SDK with Dapr and Redis.
