# Dapr Jobs Example with go-sdk

## Steps

### Prepare

- Dapr installed (v1.14 or higher)

### Run Jobs Example

<!-- STEP
name: Run Jobs Example
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
  - 'getjob - resp: Name: prod-db-backup, Schedule: @every 1s, Repeats: 10, DueTime: , TTL: , Data: value:"{\"task\":\"db-backup\",\"metadata\":{\"db_name\":\"my-prod-db\",\"backup_location\":\"/backup-dir\"}}"'
  - 'deletejob - success'

background: true
sleep: 30

-->

```bash
         dapr run --app-id=jobs \
                --metrics-port=9091 \
                --scheduler-host-address=localhost:50006 \
                --dapr-grpc-port 50001 \
                --app-port 50070 \
                --app-protocol grpc \
                --log-level debug \
                go run ./main.go

```

<!-- END_STEP -->
