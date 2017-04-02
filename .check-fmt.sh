#!/bin/sh
DIFF=$(gofmt -d -e -s .)

if [ "$DIFF" != "" ]; then
  echo "Problem with gofmt:"
  echo $DIFF
  exit 1
fi

ERRORS=$(golint server)

if [ "$ERRORS" != "" ]; then
    echo "Problem with golint:"
    echo $ERRORS
fi