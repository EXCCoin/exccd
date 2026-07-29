FROM golang:1.26-alpine3.23 AS builder
RUN apk add git ca-certificates upx gcc build-base --update --no-cache

WORKDIR /go/src/github.com/EXCCoin/exccd
COPY . .

ENV GO111MODULE=on
ARG BUILD_VERSION=2.7.6-maintenance.20260729
RUN go build -buildvcs=false -trimpath \
    -ldflags="-s -w -X github.com/EXCCoin/exccd/internal/version.Version=${BUILD_VERSION} -extldflags '-static'" .

FROM alpine:3.23

WORKDIR /app
COPY --from=builder /go/src/github.com/EXCCoin/exccd/exccd .

# Ports for the p2p of mainnet, testnet, and simnet, respectively.
EXPOSE 9666  11999  11998  11997

# Ports for the json-rpc of mainnet, testnet, and simnet, respectively.
EXPOSE 9109 19109 19556 18656

ENV DATA_DIR=/data
ENV CONFIG_FILE=/app/exccd.conf
CMD ["sh", "-c", "/app/exccd --appdata=${DATA_DIR} --configfile=${CONFIG_FILE}"]
