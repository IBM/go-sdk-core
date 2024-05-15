# Makefile to build go-sdk-core library
GO=go
LINT=golangci-lint
GOSEC=gosec
FORMATTER=goimports

COV_OPTS=-coverprofile=coverage.txt -covermode=atomic

all: tidy test lint

build:
	${GO} build ./...

testcov:
	${GO} test -tags=all ${COV_OPTS} ./...

test:
	${GO} test -tags=all ./...

lint:
	${LINT} run --build-tags=all
	DIFF=$$(${FORMATTER} -d core); if [ -n "$$DIFF" ]; then printf "\n$$DIFF\n" && exit 1; fi

scan-gosec:
	${GOSEC} ./...

format:
	${FORMATTER} -w core

tidy:
	${GO} mod tidy
