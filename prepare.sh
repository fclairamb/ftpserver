#!/bin/sh

which -s golint || go get github.com/golang/lint/golint
which -s gocyclo || go get github.com/fzipp/gocyclo
which -s goimports || go get golang.org/x/tools/cmd/goimports

go get -t -v
gofmt -e -s -l -w .
goimports -l -w .
go vet
golint ./...
gocyclo -over 15 .
go test -race ./tests
