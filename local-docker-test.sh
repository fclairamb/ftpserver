#!/bin/sh -ex

docker rm -f ftpserver ||:

GOOS=linux GOARCH=amd64 go build
docker build . -t test
if [ ! -f kitty.jpg ]; then
  curl -o kitty.jpg.tmp https://placekitten.com/2048/2048 && mv kitty.jpg.tmp kitty.jpg
fi
docker run --name ftpserver -p 2121-2130:2121-2130 test &
while ! nc -z localhost 2121 </dev/null; do sleep 1; done
curl -v -T kitty.jpg ftp://test:test@localhost:2121/
