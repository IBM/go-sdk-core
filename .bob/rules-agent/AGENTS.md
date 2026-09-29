# Agent Coding Rules

## Critical gotchas

- **Test build tags are mandatory.** Every test file has `//go:build` at line 1. Running tests without `-tags=all` (or the correct tag) produces 0 test cases and a false green. Run `cd core && go test -tags=all` or use a specific tag.
- **Formatter is `goimports`, not `gofmt`.** `make lint` diffs with `goimports -d core` — a pass with `gofmt` alone will still fail CI.
- **Do not use `== nil` on interface values.** Use `IsNil()` from `core/utils.go`; bare `== nil` misses non-nil interfaces with nil concrete values.
- **New errors must use the `IBMProblem`/`SDKProblem`/`HTTPProblem` hierarchy** (see `core/ibm_problem.go`, `core/sdk_problem.go`, `core/http_problem.go`). Plain `errors.New` / `fmt.Errorf` are not used for externally surfaced errors in this package.
- **`SDKProblem` requires the `Function` field** set to the name of the calling function — leave it blank and the computed ID will collide with other problems from the same component.
- **Apache 2.0 license header required** on every `.go` file. Match the format in any existing file.
- **Error message constants live in `core/constants.go`** (`ERRORMSG_*`). Add new ones there rather than inlining strings.
- **Version string is in `core/version.go`** (`__VERSION__`). Do not bump it manually; it is managed by `semantic-release`.
- When adding a new authenticator, register it in `core/authenticator_factory.go` — that is the single dispatch point that maps `AUTH_TYPE` string → concrete type.
