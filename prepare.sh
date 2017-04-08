#!/bin/sh

which -s golint || go get github.com/golang/lint/golint
which -s gocyclo || go get github.com/fzipp/gocyclo

go get -t -v
gofmt -e -s -l -w .
golint ./...
gocyclo -over 15 .
