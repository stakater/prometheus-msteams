FROM golang:1.25-alpine AS builder
#FROM golang:1.25 AS builder
ARG TARGETOS TARGETARCH BINARY=prometheus-msteams
#ARG TARGETOS TARGETARCH
#ARG BINARY=prometheus-msteams

WORKDIR /workspace
COPY . /workspace

RUN apk --no-cache add make bash ncurses ca-certificates tzdata && \
    update-ca-certificates && \
    CI=1 \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH} \
    make all

FROM scratch
LABEL description="A lightweight Go Web Server that accepts POST alert message from Prometheus Alertmanager and sends it to Microsoft Teams Channels using an incoming webhook url."
EXPOSE 2000
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ 
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /workspace/bin/$(BINARY) .
COPY --from=builder /workspace/default-message-workflow-card.tmpl /default-message-workflow-card.tmpl

USER 65532:65532

ENTRYPOINT ["/$(BINARY)"]
