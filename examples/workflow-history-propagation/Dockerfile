# The binary is expected to be built on the host and copied in.
# Build this from examples/workflow-history-propagation, where examples/go.mod
# provides the local durabletask-go / go-sdk paths via replace directives.
# Build with:
#   CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o payment-app .
# (use GOARCH=amd64 on Intel machines)
FROM gcr.io/distroless/base-debian12
COPY payment-app /payment-app
ENTRYPOINT ["/payment-app"]
