PLATFORM=$(shell uname -s | tr '[:upper:]' '[:lower:]')
VERSION := $(shell grep -Eo '((v|PR)[0-9]+[\.][0-9]+[\.][0-9]+(-[a-zA-Z0-9]*)?)' version.go)
COMMIT_HASH :=$(shell git rev-parse --short HEAD)

USERID := $(shell id -u $$USER)
GROUPID:= $(shell id -g $$USER)

export GOPRIVATE=github.com/moovfinancial

# Common CI/CD commands
update:
	go mod vendor
	# pkger -include /migrations -include /configs/config.default.yml

build:
	go build -mod=vendor -o bin/server github.com/moovfinancial/ftpserver/

.PHONY: setup
setup:
	-docker-compose up -d --remove-orphans --force-recreate
	sleep 10

.PHONY: check
check: setup
ifeq ($(OS),Windows_NT)
	@echo "Skipping checks on Windows, currently unsupported."
else
	@wget -O lint-project.sh https://raw.githubusercontent.com/moov-io/infra/master/go/lint-project.sh
	@chmod +x ./lint-project.sh
	GOCYCLO_LIMIT=27 ./lint-project.sh
endif

docker: update build
	docker build --pull -t moovfinancial/ftpserver:$(VERSION) -f Dockerfile .

.PHONY: docker-push
docker-push:
	docker push moovfinancial/ftpserver:$(VERSION)

.PHONY: docker-dev
docker-dev: update build
	docker build --pull -t moovfinancial/ftpserver:dev-$(COMMIT_HASH) -f Dockerfile .

.PHONY: docker-push-dev
docker-push-dev:
	docker push moovfinancial/ftpserver:dev-$(COMMIT_HASH)

.PHONY: teardown
teardown:
	-docker-compose down --remove-orphans