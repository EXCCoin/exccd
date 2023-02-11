#!/bin/sh

set -ex

# The script does automatic checking on a Go package and its sub-packages,
# including:
# 1. gofmt         (https://golang.org/cmd/gofmt/)
# 2. gosimple      (https://github.com/dominikh/go-simple)
# 3. unconvert     (https://github.com/mdempsky/unconvert)
# 4. ineffassign   (https://github.com/gordonklaus/ineffassign)
# 5. go vet        (https://golang.org/cmd/vet)
# 6. misspell      (https://github.com/client9/misspell)

# golangci-lint (github.com/golangci/golangci-lint) is used to run each
# static checker.

go version

go test -vet=all -short ./...

# run linters
golangci-lint run --build-tags=rpctest --disable-all --deadline=10m \
  --enable=gofmt \
  --enable=gosimple \
  --enable=unconvert \
  --enable=ineffassign \
  --enable=govet \
  --enable=misspell \
  --enable=unused

echo "------------------------------------------"
echo "Tests completed successfully!"
