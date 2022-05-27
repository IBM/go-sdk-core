# Makefile to build go-sdk-core library
GO=go
LINT=golangci-lint
GOSEC=gosec

VDIR=v5

COV_OPTS=-coverprofile=coverage.txt -covermode=atomic

all: testcov lint tidy

testcov:
	cd ${VDIR} && ${GO} test -tags=all ${COV_OPTS} ./...

test:
	cd ${VDIR} && ${GO} test -tags=all ./...

lint:
	cd ${VDIR} && ${LINT} run --build-tags=all

scan-gosec:
	cd ${VDIR} && ${GOSEC} ./...

tidy:
	cd ${VDIR} && ${GO} mod tidy
