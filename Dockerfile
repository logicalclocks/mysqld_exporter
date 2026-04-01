ARG TARGETOS
ARG TARGETARCH
FROM --platform=$BUILDPLATFORM golang:1.12 AS builder

WORKDIR /go/src/github.com/prometheus/mysqld_exporter

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN GIT_REVISION=$(git rev-parse HEAD 2>/dev/null || echo "unknown") && \
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown") && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=vendor -a -tags netgo \
    -ldflags "-X github.com/prometheus/common/version.Version=$(cat VERSION) \
              -X github.com/prometheus/common/version.Revision=${GIT_REVISION} \
              -X github.com/prometheus/common/version.Branch=${GIT_BRANCH} \
              -X github.com/prometheus/common/version.BuildDate=$(date +%Y%m%d-%H:%M:%S)" \
    -o /mysqld_exporter .

FROM        quay.io/prometheus/busybox-${TARGETOS}-${TARGETARCH}:latest

COPY --from=builder /mysqld_exporter /bin/mysqld_exporter

USER        nobody
EXPOSE      9104
ENTRYPOINT  [ "/bin/mysqld_exporter" ]
