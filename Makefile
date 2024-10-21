# Makefile to build go-sdk-core library
GO=go
LINT=golangci-lint
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

format:
	${FORMATTER} -w core

tidy:
	${GO} mod tidy

detect-secrets:
	detect-secrets scan --update .secrets.baseline
	detect-secrets audit .secrets.baseline
