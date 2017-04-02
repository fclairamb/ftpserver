#!/bin/sh -e
DIFF=$(gofmt -d -e -s .)

if [ "$DIFF" != "" ]; then
  echo "Problem with gofmt:"
  echo $DIFF
  exit 1
fi

golint -set_exit_status=1 ./...

