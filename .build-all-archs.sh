#!/usr/bin/env bash

version=$(go version|grep -Eo go[0-9]\.[0-9]+)

if [ "$version" != "go1.13" ]; then
    echo "Docker images are only generated for Go 1.13 and you have ${version}."
    exit 0
fi

for params in "GOOS=linux GOARCH=amd64 EXT=" "GOOS=linux GOARCH=386" "GOOS=linux GOARCH=arm" "GOOS=darwin GOARCH=amd64" "GOOS=windows GOARCH=amd64 EXT=.exe" "GOOS=windows GOARCH=386"
do
  export $params
  echo -n "Building for OS=${GOOS} and ARCH=${GOARCH} ."
  go get && echo -n "."
  go build -o ftpserver-${GOOS}-${GOARCH}${EXT}
  echo .
done
