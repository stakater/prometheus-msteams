FROM golang:1.25-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG COMMIT
ARG BRANCH
ARG BUILD_DATE

WORKDIR /workspace
COPY . /workspace

RUN apk --no-cache add make git bash ncurses ca-certificates tzdata && \
    update-ca-certificates && \
    CI=1 CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    VERSION=${VERSION:-dev} COMMIT=${COMMIT:-dirty} BRANCH=${BRANCH:-main} \
    BUILD_DATE=${BUILD_DATE:-"1970-01-01T00:00:00Z"} BINDIR=/workspace/bin \
    make build && \
    ls -l /workspace/bin
    #make all && \
    
FROM scratch
LABEL description="A lightweight Go Web Server that accepts POST alert message from Prometheus Alertmanager and sends it to Microsoft Teams Channels using an incoming webhook url."
EXPOSE 2000
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ 
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /workspace/default-message-workflow-card.tmpl /default-message-workflow-card.tmpl
COPY --from=builder /workspace/bin/prometheus-msteams .

USER 65532:65532

ENTRYPOINT ["/prometheus-msteams"]
