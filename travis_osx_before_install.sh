#!/bin/bash

export DEP_TAG='v0.4.1'
export GLIDE_TAG='v0.13.1'
export GOMETALINTER_TAG='v2.0.5'

go get -v github.com/alecthomas/gometalinter         && \
cd $GOPATH/src/github.com/alecthomas/gometalinter    && \
git checkout $GOMETALINTER_TAG                       && \
go install                                           && \
gometalinter --install
cd -

# go get -u honnef.co/go/tools/... ?
go get -v github.com/Masterminds/glide               && \
cd $GOPATH/src/github.com/Masterminds/glide          && \
git checkout $GLIDE_TAG                              && \
make build                                           && \
mv glide `which glide`                               && \
cd -
    
go get -v github.com/golang/dep                      && \
cd $GOPATH/src/github.com/golang/dep                 && \
git checkout $DEP_TAG                                && \
go install ./...                                     && \
cd -
