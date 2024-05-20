# Dapr Distributed Scheduler Example with go-sdk

## Steps

### Prepare

- Dapr installed (v1.14<)

### Run scheduler

This step needs to be removed

<!-- STEP
name: Run scheduler
output_match_mode: substring
expected_stdout_lines:
  - 'Dapr Scheduler listening on: 127.0.0.1:50006'

background: true
sleep: 180
-->

```bash
        ~/.dapr/bin/scheduler
```

<!-- END_STEP -->

### Run new dapr sidecar with scheduler

This step needs to be removed

<!-- STEP
name: Run sidecar
output_match_mode: substring
expected_stdout_lines:
  - 'Scheduler stream connected'

background: true
sleep: 90
-->

```bash
        ~/.dapr/bin/daprd --app-id=distributed-scheduler \
                --metrics-port=9091 \
                --scheduler-host-address=127.0.0.1:50006 \
                --dapr-grpc-port 50001 \
                --app-port 50070 \
                --app-protocol grpc \
                --log-level debug
```

<!-- END_STEP -->

### Run Distributed Scheduling Example

<!-- STEP
name: Run Distributed Scheduling Example
output_match_mode: substring
expected_stdout_lines:
  - 'schedulejob - success'
  - 'job 0 received'
  - 'extracted payload: {db-backup {my-prod-db /backup-dir}}'
  - 'job 1 received'
  - 'extracted payload: {db-backup {my-prod-db /backup-dir}}'
  - 'job 2 received'
  - 'extracted payload: {db-backup {my-prod-db /backup-dir}}'
  - 'getjob - resp: &{prod-db-backup @every 1s 10   value:"{\"task\":\"db-backup\",\"metadata\":{\"db_name\":\"my-prod-db\",\"backup_location\":\"/backup-dir\"}}"}'
  - 'deletejob - success'


background: false
sleep: 60
-->

```bash
         go run ./main.go
```

<!-- END_STEP -->



