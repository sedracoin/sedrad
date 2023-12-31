#!/bin/sh -ex

FLAGS=$@

go version

go get $FLAGS -t -d ./...
# This is to bypass a go bug: https://github.com/golang/go/issues/27643
GO111MODULE=off go get $FLAGS golang.org/x/lint/golint \
                          honnef.co/go/tools/cmd/staticcheck

test -z "$(go fmt ./...)"

golint -set_exit_status ./...

staticcheck -checks=all -tags=gofuzz ./...

go vet $FLAGS -tags=gofuzz ./...

go build $FLAGS .

go test $FLAGS -tags=gofuzz ./...