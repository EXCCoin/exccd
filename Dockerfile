FROM golang:1.20-alpine3.18 AS builder
RUN apk add git ca-certificates upx gcc build-base --update --no-cache

WORKDIR /go/src/github.com/EXCCoin/exccd
COPY . .

ENV GO111MODULE=on
RUN go build -ldflags='-s -w -X main.appBuild=alpine:3.18 -extldflags "-static"' .

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /go/src/github.com/EXCCoin/exccd/exccd .

# Ports for the p2p of mainnet, testnet, and simnet, respectively.
EXPOSE 9666  11999  11998  11997

# Ports for the json-rpc of mainnet, testnet, and simnet, respectively.
EXPOSE 9109 19109 19556 18656

ENV DATA_DIR=/data
ENV CONFIG_FILE=/app/exccd.conf
CMD ["sh", "-c", "/app/exccd --appdata=${DATA_DIR} --configfile=${CONFIG_FILE}"]
