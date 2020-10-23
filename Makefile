# Makefile to build go-sdk-core library

VDIR=v4

all: build test lint tidy

build:
	cd ${VDIR} && go build ./...

test:
	cd ${VDIR} && go test -tags=all ./...

lint:
	cd ${VDIR} && golangci-lint run --build-tags=all

tidy:
	cd ${VDIR} && go mod tidy
