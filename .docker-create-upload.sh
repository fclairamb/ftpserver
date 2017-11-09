#!/bin/bash -ex

version=$(go version|grep -Eo go[0-9]\.[0-9])

if [ "$version" != "go1.9" ]; then
    echo "Docker images are only generated for Go 1.9 and you have ${version}."
    exit 0
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo
# GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -a -installsuffix cgo

echo "Docker repo: ${DOCKER_REPO}:${TRAVIS_COMMIT}"

DOCKER_NAME=${DOCKER_REPO}:${TRAVIS_COMMIT}

docker build -t ${DOCKER_NAME} .

docker tag ${DOCKER_NAME} ${DOCKER_REPO}:travis-${TRAVIS_BUILD_NUMBER}

if [[ "${TRAVIS_TAG}" = "" ]]; then
    if [[ "${TRAVIS_BRANCH}" = "master" ]]; then
        DOCKER_TAG=latest
    else
        DOCKER_TAG=${TRAVIS_BRANCH}
    fi
else
    DOCKER_TAG=${TRAVIS_TAG}
fi

docker tag ${DOCKER_NAME} ${DOCKER_REPO}:${DOCKER_TAG}

# If you execute locally:
# docker rm -f ftpserver 2>/dev/null ||:

# Let's check that the container is actually fully usable
docker run -d -p 2121-2200:2121-2200 --name=ftpserver ${DOCKER_NAME}

# We wait for the server to be ready
for (( i=0; i < 30; i++))
do
  out=$(echo "QUIT" | nc localhost 2121 -w 1)
  if [[ "${out}" == *"220 "* ]]; then
    break
  fi
  sleep 1
done

# Checking that by default the localpath is the "/data" directory
path=$(curl -s ftp://test:test@localhost:2121/virtual/localpath.txt)
if [ "${path}" != "/data/shared" ]; then
    echo "The path is wrong: ${path}"
    exit 1
fi

# Checking that upload/download is working fine
chk_before=$(shasum ftpserver| cut -d " " -f 1)
curl -s -T ftpserver ftp://test:test@localhost:2121/upload
curl -s -o ftpserver_downloaded ftp://test:test@localhost:2121/upload
chk_after=$(shasum ftpserver_downloaded| cut -d " " -f 1)
if [ "${chk_before}" != "${chk_after}" ]; then
    echo "Checksum mismatch"
    exit 1
fi

# Check the file listing is working fine
curl -s ftp://test:test@localhost:2121/

if [[ "${DOCKER_PASSWORD}" = "" ]]; then
    echo "Probably a PR"
    exit 0
fi

# florent(2017-10-27): Issue 47: Pull requests should pass tests
docker login -u="${DOCKER_USERNAME}" -p="${DOCKER_PASSWORD}"

# florent(2017-10-27): Docker hub is becoming dirty. Let's only keep the branches and tags
docker push ${DOCKER_REPO}:${DOCKER_TAG}
