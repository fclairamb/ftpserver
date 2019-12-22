#!/bin/sh -e
DIFF=$(gofmt -d -e -s .)

version=$(go version|grep -Eo go[0-9]\.[0-9]+)

if [ "$version" != "go1.13" ]; then
    echo "Skipping go fmt for ${version}"
    exit 0
fi

if [ "$DIFF" != "" ]; then
  echo "Problem with gofmt:"
  echo $DIFF
  exit 1
fi

golint -set_exit_status=1 ./...

