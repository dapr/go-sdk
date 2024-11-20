# Dapr Distributed Scheduler Example with go-sdk

## Steps

### Prepare

- Dapr installed (v1.14 or higher)

### Run Distributed Scheduling Example

<!-- STEP
name: Run Distributed Scheduling Example
output_match_mode: substring
expected_stdout_lines:
  - 'Scheduler stream connected'
  - 'schedulejob - success'
  - 'job 0 received'
  - 'payload: {db-backup {my-prod-db /backup-dir}}'
  - 'job 1 received'
  - 'payload: {db-backup {my-prod-db /backup-dir}}'
  - 'job 2 received'
  - 'payload: {db-backup {my-prod-db /backup-dir}}'
  - 'getjob - resp: &{prod-db-backup @every 1s 10   value:"{\"task\":\"db-backup\",\"metadata\":{\"db_name\":\"my-prod-db\",\"backup_location\":\"/backup-dir\"}}"}'
  - 'deletejob - success'

background: true
sleep: 30

-->

```bash
         dapr run --app-id=distributed-scheduler \
                --metrics-port=9091 \
                --scheduler-host-address=localhost:50006 \
                --dapr-grpc-port 50001 \
                --app-port 50070 \
                --app-protocol grpc \
                --log-level debug \
                go run ./main.go

```

<!-- END_STEP -->
