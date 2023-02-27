
# TODO: Needs some documentation here about the RPC server, logging via docker
# logs, mounting a volume, the conf file, etc...

###############
# Builder Stage
###############

# Basic Go environment with git, SSL CA certs, and upx.
# golang:1.17.1-alpine (linux/amd64)
FROM golang@sha256:13919fb9091f6667cb375d5fdf016ecd6d3a5d5995603000d422b04583de4ef9 AS builder
RUN apk add --no-cache git ca-certificates upx

# Empty directory to be copied into place in the production image since it will
# run as a non-root container and thus not have permissions to create
# directories or change ownership of anything outside of the structure already
# created for it.
RUN mkdir /emptydatadir

# New unprivileged user for use in production image below to improve security.
ENV USER=excc
ENV UID=10000
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home="/home/${USER}" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Build dcrd and other commands it provides
WORKDIR /go/src/github.com/EXCCCoin/exccd
RUN git clone https://github.com/EXCCCoin/exccd . && \
    CGO_ENABLED=0 GOOS=linux \
    go install -trimpath -tags safe,netgo,timetzdata \
      -ldflags="-s -w" \
      . ./cmd/gencerts ./cmd/promptsecret

# Build dcrctl
WORKDIR /go/src/github.com/EXCCCoin/exccctl
RUN git clone https://github.com/EXCCoin/exccctl . && \
    CGO_ENABLED=0 GOOS=linux \
      go install -trimpath -tags safe,netgo -ldflags="-s -w"

# Build entrypoint helper for the production image.
WORKDIR /go/src/github.com/EXCCoin/exccd/contrib/docker/entrypoint
COPY ./contrib/docker/entrypoint/entrypoint.go .
RUN go mod init entrypoint && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux \
    go install -trimpath -tags netgo,timetzdata -ldflags="-s -w" .

# Compress bins
RUN upx -9 /go/bin/*

##################
# Production image
##################

# Minimal scratch-based environment.
FROM scratch
ENV EXCC_DATA=/home/excc
#ENV EXCCD_EXPOSE_RPC=false # TODO: Want something like this?
ENV EXCCD_NO_FILE_LOGGING=true
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/* /bin/
COPY --from=builder --chown=excc /emptydatadir /tmp

# Use an unprivileged user.
USER excc

# Ports for the p2p of mainnet, testnet, and simnet, respectively.
EXPOSE 9666  11999  11998  11997

# Ports for the json-rpc of mainnet, testnet, and simnet, respectively.
EXPOSE 9109 19109 19556 18656

ENTRYPOINT [ "/bin/entrypoint" ]

RUN [ "exccd", "--version" ]

# TODO: Want this or not?  I've seen conflicting info and I'm not a docker expert...
#VOLUME [ "/home/excc" ]
