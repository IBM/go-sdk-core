# Makefile to build go-sdk-core library
GO=go
LINT=golangci-lint
GOSEC=gosec

COV_OPTS=-coverprofile=coverage.txt -covermode=atomic

all: tidy test lint

testcov:
	${GO} test -tags=all ${COV_OPTS} ./...

test:
	${GO} test -tags=all ./...

lint:
	${LINT} run --build-tags=all

scan-gosec:
	${GOSEC} ./...

tidy:
	${GO} mod tidy
