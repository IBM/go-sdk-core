# Plan / Architecture Rules

## Architectural constraints

- **Single flat package.** All code is `package core` in `core/`. Do not introduce sub-packages — generated SDK clients import the package as a unit.
- **Authenticator interface is the extension point.** New auth strategies must implement `AuthenticationType() string`, `Authenticate(*http.Request) error`, and `Validate() error` (see `core/authenticator.go`). They must also be registered in `core/authenticator_factory.go`.
- **Token caching is per-authenticator.** IAM, CP4D, VPC, MCSP, and Container authenticators each manage their own cached token with a mutex (`sync.Once` / `sync.Mutex`). Any new token-based authenticator must follow the same pattern to be thread-safe.
- **`BaseService` owns the retry layer.** Retries are handled via `hashicorp/go-retryablehttp` and controlled by `ServiceOptions` fields (`EnableRetries`, `MaxRetries`, `RetryMaxInterval`). Do not add retry logic inside authenticators.
- **Problem IDs are deterministic hashes.** `GetID()` on `SDKProblem`/`HTTPProblem` computes a hash from component + function/operationID + status code. Changing the signature fields is a breaking change for consumers who key on problem IDs.
- **Build tags partition the test suite.** Tags (`all`, `fast`, `slow`, `auth`, `basesvc`, `retries`, `log`) are the only mechanism to group tests. CI runs `-tags=all`; individual tag runs are used for speed. Plan new test files with an appropriate tag set.
- **`semantic-release` drives versioning.** Commit message format (Angular) is not optional — it is the version-bump input. Breaking changes require a `BREAKING CHANGE:` footer.
