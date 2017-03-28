#!/bin/sh
DIFF=$(gofmt -d -e -s .)

if [ "$DIFF" != "" ]; then
  echo "Problem with format: $DIFF"
  exit 1
fi
