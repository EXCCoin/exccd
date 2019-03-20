FROM golang:1.12.1-alpine3.9 as builder

RUN apk add git gcc g++ musl-dev --update --no-cache

WORKDIR /go/src/github.com/EXCCoin/exccd
COPY . .

ENV GO111MODULE=on
RUN go build -ldflags='-s -w -X main.appBuild=alpine3.9 -extldflags "-static"' .


FROM alpine:3.9

WORKDIR /app
COPY --from=builder /go/src/github.com/EXCCoin/exccd/exccd .

EXPOSE 9666
ENV DATA_DIR=/data
ENV CONFIG_FILE=/app/exccd.conf
CMD ["sh", "-c", "/app/exccd --appdata=${DATA_DIR} --configfile=${CONFIG_FILE}"]
