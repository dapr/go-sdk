FROM alpine
WORKDIR /app
ADD build build
ENTRYPOINT build/server-test
CMD go test build/server-test
