#!/bin/sh -ex

docker rm -f ftpserver ||:

GOOS=linux GOARCH=amd64 go build
docker build . -t test
docker run --name ftpserver -p 2121:2121 -p 2122:2122 -p 2123:2123 test -conf /etc/ftpserver_test.toml &
sleep 1
curl -v -T main.go ftp://test:test@localhost:2121/
