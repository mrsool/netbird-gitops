FROM golang:1.22.2 AS builder

COPY . /go/src/github.com/instabug/netbird-gitops/
WORKDIR /go/src/github.com/instabug/netbird-gitops/
RUN set -Eeux && \
    go mod download && \
    go mod verify

RUN GOOS=linux GOARCH=amd64 \
    go build \
    -o app cmd/

FROM alpine:3.17.1
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/instabug/netbird-gitops/app .

EXPOSE 8123
ENTRYPOINT ["./app"]