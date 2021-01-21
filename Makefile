# Makefile to build go-sdk-core library

VDIR=v5

COV_OPTS=-coverprofile=coverage.txt -covermode=atomic

all: testcov lint tidy

testcov:
	cd ${VDIR} && go test -tags=all ${COV_OPTS} ./...

test:
	cd ${VDIR} && go test -tags=all ./...

lint:
	cd ${VDIR} && golangci-lint run --build-tags=all

tidy:
	cd ${VDIR} && go mod tidy
