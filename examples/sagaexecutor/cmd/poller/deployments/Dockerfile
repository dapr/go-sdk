FROM golang:1.18-alpine as golang
RUN apk --update add ca-certificates

WORKDIR /app

FROM scratch

COPY --from=golang /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=golang /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=golang /etc/passwd /etc/passwd
COPY --from=golang /etc/group /etc/group

ADD build build
ENTRYPOINT build/poller

