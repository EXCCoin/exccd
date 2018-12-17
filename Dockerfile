FROM golang:1.11

RUN apt-get update && \
    apt-get install -qy rsync && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/EXCCoin/exccd
COPY . .

ENV GO111MODULE=on
RUN go install . ./cmd/...

EXPOSE 9666

CMD [ "exccd" ]
