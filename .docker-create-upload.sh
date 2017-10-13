#!/bin/sh -e

version=$(go version|grep -Eo go[0-9\.]+)

if [ "$version" != "go1.9" ]; then
    echo "Container are only generated for version 1.9 and you have ${version}."
    exit 0
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo
# GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -a -installsuffix cgo

docker login -u="${DOCKER_USERNAME}" -p="${DOCKER_PASSWORD}"

echo "Docker repo: ${DOCKER_REPO}:${TRAVIS_COMMIT}"

DOCKER_NAME=${DOCKER_REPO}:${TRAVIS_COMMIT}

docker build -t ${DOCKER_NAME} .

docker tag ${DOCKER_NAME} ${DOCKER_REPO}:travis-${TRAVIS_BUILD_NUMBER}

if [ "${TRAVIS_TAG}" = "" ]; then
    if [ "${TRAVIS_BRANCH}" = "master" ]; then
        DOCKER_TAG=latest
    else
        DOCKER_TAG=${TRAVIS_BRANCH}
    fi
else
    DOCKER_TAG=${TRAVIS_TAG}
fi

docker tag ${DOCKER_NAME} ${DOCKER_REPO}:${DOCKER_TAG}

docker push ${DOCKER_REPO}

#docker run -ti ${DOCKER_NAME}
