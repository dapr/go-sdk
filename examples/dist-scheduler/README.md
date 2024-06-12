# Dapr Distributed Scheduler Example with go-sdk

## Steps

### Prepare

- Dapr installed (v1.14<)

### Run Distributed Scheduling Example

<!-- STEP
name: Run sidecar
output_match_mode: substring
expected_stdout_lines:
  - 'Scheduler stream connected'
  - 'schedulejob - success'
  - 'job 0 received'
  - 'extracted payload: {db-backup {my-prod-db /backup-dir}}'
  - 'job 1 received'
  - 'extracted payload: {db-backup {my-prod-db /backup-dir}}'
  - 'job 2 received'
  - 'extracted payload: {db-backup {my-prod-db /backup-dir}}'
  - 'getjob - resp: &{prod-db-backup @every 1s 10   value:"{\"task\":\"db-backup\",\"metadata\":{\"db_name\":\"my-prod-db\",\"backup_location\":\"/backup-dir\"}}"}'
  - 'deletejob - success'

background: true
sleep: 60
-->

```bash
        dapr run \
                --app-id=distributed-scheduler \
                --metrics-port=9091 \
                --dapr-grpc-port 50001 \
                --app-port 50070 \
                --app-protocol grpc \
                --log-level debug \
                -- go run .
```

<!-- END_STEP -->

