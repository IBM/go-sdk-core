# Ask Mode Context

- All user-facing code lives in `core/` — there are no subdirectories; the package is flat.
- "Services" in this repo means the `BaseService` struct consumed by generated SDK clients, not standalone microservices.
- `Authentication.md` in the project root is the canonical auth reference; the `README.md` points there.
- Config property names (`PROPNAME_*` constants in `core/constants.go`) are what end-users set as environment variables with the pattern `<SERVICE_NAME>_<PROPERTY>` (e.g. `MYSERVICE_APIKEY`).
- `UnmarshalPrimitive` / `UnmarshalModel` in `core/unmarshal_v2.go` are the deserialization helpers used by generated SDK code — not standard `json.Unmarshal`.
- The `Problem` / `IBMProblem` / `SDKProblem` / `HTTPProblem` type chain (introduced in 2024) is the structured error system; older code may still use plain errors in tests.
