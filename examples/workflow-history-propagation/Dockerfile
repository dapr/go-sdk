# The binary is expected to be built on the host (which has access to the
# go.work workspace with local durabletask-go / go-sdk paths) and copied in.
# Build with:
#   CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o payment-app .
# (use GOARCH=amd64 on Intel machines)
FROM gcr.io/distroless/base-debian12
COPY payment-app /payment-app
ENTRYPOINT ["/payment-app"]
