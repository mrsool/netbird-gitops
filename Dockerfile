FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.1 AS builder

COPY . /go/src/github.com/mrsool/netbird-gitops/
WORKDIR /go/src/github.com/mrsool/netbird-gitops/
RUN set -Eeux && \
    go mod download && \
    go mod verify

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -o app cmd/*

FROM alpine:3.20.3
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/mrsool/netbird-gitops/app .

ENTRYPOINT ["/root/app"]
