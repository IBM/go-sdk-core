# Makefile to build go-sdk-core library

VDIR=v5

all: build test lint tidy

build:
	cd ${VDIR} && go build ./...

test:
	cd ${VDIR} && go test ./...

lint:
	cd ${VDIR} && golangci-lint run

tidy:
	cd ${VDIR} && go mod tidy
