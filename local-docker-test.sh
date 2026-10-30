#!/bin/sh -ex
# Builds the docker image and checks that a file can be uploaded and downloaded back

docker rm -f ftpserver-test || :

docker build . -t ftpserver-test
docker run -d --rm --name ftpserver-test -p 2121-2130:2121-2130 ftpserver-test
trap 'docker rm -f ftpserver-test' EXIT

# The auto-created config file gets a random password, printed in the logs
timeout 30 sh -c 'until docker logs ftpserver-test 2>&1 | grep -q "No conf file"; do sleep 1; done'
PASS="$(docker logs ftpserver-test 2>&1 | grep "No conf file" | sed -E 's/.*password=([^ ]+).*/\1/')"

head -c 2097152 /dev/urandom > local.bin.tmp
while ! nc -z localhost 2121 </dev/null; do sleep 1; done
curl -v -T local.bin.tmp "ftp://test:${PASS}@localhost:2121/remote.bin"
curl -v "ftp://test:${PASS}@localhost:2121/remote.bin" -o remote.bin.tmp
cmp local.bin.tmp remote.bin.tmp
rm local.bin.tmp remote.bin.tmp
