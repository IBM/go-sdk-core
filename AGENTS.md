# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Stack

Go SDK core library (`github.com/IBM/go-sdk-core/v5`). All source lives in `core/`. No subdirectories — flat package.

## Commands

```sh
make all          # tidy + test + lint (standard CI gate)
make test         # go test -tags=all ./...
make lint         # golangci-lint + goimports diff check on core/
make format       # goimports -w core   (formatter is goimports, NOT gofmt alone)
make tidy         # go mod tidy
```

### Run a single test file / tag group (must cd into `core/` first)

```sh
cd core && go test -tags=basesvc        # basesvc, retries, auth, log, fast, slow, all
cd core && go test -run TestFooBar      # by test name (requires -tags=all or matching tag)
```

**Critical**: every test file has a `//go:build` constraint at line 1 (e.g. `//go:build all || fast || basesvc`). Running `go test` without `-tags=all` from the repo root skips most tests. Always include `-tags=all` (or the specific tag) or tests silently pass with 0 cases.

## Test Framework

Uses **both** Ginkgo/Gomega (`core_suite_test.go` bootstraps the suite) and standard `testify` assertions. New tests should match the style of the file they extend.

Internal tests (package `core`) and external tests (package `core_test`) coexist — `common_test.go` is `package core` (white-box) while `core_suite_test.go` is `package core_test` (black-box). Follow the existing package declaration of the file you're editing.

## Error Handling

The project has a custom error hierarchy — **do not use plain `errors.New` or `fmt.Errorf` for new errors in this package**:
- `IBMProblem` — base struct (embed it)
- `SDKProblem` — for errors originating in SDK code; requires `Function` field (caller name)
- `HTTPProblem` — for failed HTTP responses; carries `DetailedResponse`

Use `NewProblemComponent(name, version)` to supply the `Component` field. IDs are computed deterministically from fields; the `discriminator` string disambiguates otherwise-identical problems.

## Code Style

- Formatter: **`goimports`** (not `gofmt`). `make lint` will fail if `goimports -d core` produces a diff.
- Linter: `golangci-lint` with `gosec` enabled. `staticcheck` runs with `all` minus `ST1005`/`ST1003` (error string capitalisation and naming checks are disabled).
- Constants that map to external config property names use `SCREAMING_SNAKE_CASE` (e.g. `PROPNAME_APIKEY`).
- Auth type string constants (`AUTHTYPE_*`) are lower-camel (`"bearerToken"`, `"iamAssume"`) — matching is case-insensitive in config parsing.
- `IsNil()` in `utils.go` is the canonical nil-check for interface/pointer types; use it instead of `== nil` for interface values.
- License header (Apache 2.0, IBM copyright) is required on every `.go` file.

## Commit Messages

Must follow **Angular commit message format** — `semantic-release` uses this to determine version bumps and generate changelogs. Example: `fix(IamAuthenticator): handle token expiry edge case`.
