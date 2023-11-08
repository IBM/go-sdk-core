# Makefile to build go-sdk-core library
GO=go
LINT=golangci-lint
GOSEC=gosec
FORMATTER=goimports

COV_OPTS=-coverprofile=coverage.txt -covermode=atomic

all: tidy test lint

testcov:
	${GO} test -tags=all ${COV_OPTS} ./...

test:
	${GO} test -tags=all ./...

lint:
	${LINT} run --build-tags=all
	${FORMATTER} -d core
	if [[ -n `${FORMATTER} -d core` ]]; then exit 1; fi

scan-gosec:
	${GOSEC} ./...

format:
	${FORMATTER} -w core

tidy:
	${GO} mod tidy
