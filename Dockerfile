FROM golang:1.10

ENV TERM linux
ENV USER node

# create user
RUN adduser --disabled-password --gecos ''  $USER

# update base distro & install build tooling
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get install -qy rsync

# create directory for build artifacts, adjust user permissions
RUN mkdir /release && \
    chown $USER /release

# switch user
USER $USER
ENV HOME /home/$USER

#Get deps
ENV DEP_TAG v0.4.1

RUN go get -v github.com/golang/dep && \
    cd /go/src/github.com/golang/dep && \
    git checkout $DEP_TAG && \
    go install ./...

#Get exccd
ENV EXCCD_BRANCH task/DEX-161_test_mainnet

RUN go get -v github.com/EXCCoin/exccd && \
    cd /go/src/github.com/EXCCoin/exccd && \
    git checkout $EXCCD_BRANCH && \
    dep ensure && \
    go install ./...