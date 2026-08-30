# Tests, Generated Code, and Verification

## Test Location and Style

- Unit tests tightly coupled to a package live beside the production files in the same package and use the `*_test.go` suffix. Runtime-facing or cross-runtime E2E tests live under that runtime's `test/e2e/`. For example, HTTP behavior for a status endpoint belongs at `runtimes/<runtime>/test/e2e/status/status_test.go`, not under `transports/status/`; cross-runtime integration and architecture tests live in the independent `test/` module.
- `test/e2e/` verifies only cross-package flows, HTTP/WebSocket boundaries, or complete runtime-to-runtime flows; it does not hold ordinary unit tests. Pure logic for one utility, parser, or service stays beside its production package, while tests that verify an exposed endpoint through HTTP, Gin, ServeMux, or WebSocket belong to the runtime's E2E suite.
- `contracts/`, `shared/`, and each runtime keep only their own unit tests. `test/go.mod` manages cross-module test dependencies and local `replace` directives; those tests must not depend on the root module.
- When adding or changing a Gateway controller, use `httptest` and a Gin test router to verify request parsing, actor/context fields, and invalid input.
- Service/repository tests should first cover high-risk behavior such as permissions, soft delete, transactions, and error branches rather than creating template tests for every getter.
- Put reusable JSON fixtures under the relevant domain's `testdata/` and name them with descriptive `snake_case.json` filenames.
- Tests must be independent and repeatable; they must not depend on execution order, the wall clock, external networks, or shared production data.

## Generated Code and Contracts

- `contracts/core/v1/graphql/generated` and `contracts/core/v1/graphql/models` are generated GraphQL artifacts. After changing `contracts/core/v1/graphql` sources, run `make -C contracts gql-generate` or `make -C contracts gql-regenerate`; never edit generated code by hand. Keep gqlgen generated output in these contracts paths.
- Put public API route semantics in `docs/api-route-design/`, code and data-model design in `docs/codebase-design/`, and Realtime, Yjs, and cross-runtime protocols in `docs/system-design/`. A change to public semantics must update the corresponding design document and tests in the same change.
- `infra/` contains deployment and observability configuration. After changes, verify that Docker Compose, Nginx, and OTEL/Grafana settings remain consistent.

## Go Modules and Test Entry Points

- The repository root keeps only `go.work` as the workspace manifest; it does not keep a root `go.mod`/`go.sum`.
- `make test-all` uses `runtimes/cli` to run contracts, shared, each Go runtime, and the root `test/` module in order. To test one module, use `make test-module MODULE=<module>` or run `GOWORK=off go test ./...` directly in that module.
- `make test-architecture` and `make test-integration` run only the corresponding layer in the root `test/` module; runtime E2E tests are run by the owning runtime's Makefile.
- Root integration tests use the committed `infra/docker/docker-compose.integration.yaml` for PostgreSQL, Redis, and Kafka. Local development and GitHub Actions must use the same Compose stack; the test code must not create Testcontainers itself. Root `compose-integration-up` and `compose-integration-down` own infrastructure lifecycle; `make test-integration` and `make test-integration-kafka` only execute tests and can reuse already-running services. `make test-integration-managed` provides a one-command local start/test/cleanup lifecycle; CI keeps setup, test, and cleanup explicit. Test data belongs in `test/integration/testdata/`.
- WebSocket load/soak and Kafka consumer-lag verification use Grafana k6 scripts under `test/load/`, invoked through `make test-load-websocket`, `make test-soak-websocket`, and `make test-load-kafka-lag`; these scripts are not part of any Go module.
- GraphQL generation commands are maintained by `contracts/Makefile` and run from the contracts module with `make -C contracts gql-generate`; artifacts remain under `contracts/core/v1/graphql/`. Do not restore a root `go.mod` for tooling commands.
- The root Makefile provides a stable cross-environment entry point and owns the integration Compose lifecycle. Module-specific commands live in the Makefiles under `contracts/`, `shared/`, each runtime, and `test/`. `runtimes/cli` uses Cobra and subprocesses to invoke those module commands and does not import test packages directly. `test/Makefile` integration targets only execute tests; they do not start or stop Docker.

## Minimum Verification for Every Change

1. Run `gofmt` on modified Go files.
2. Run `go test` for affected packages; when a safe integration/E2E environment is available, also run the corresponding E2E test.
3. When changing a schema, generating GraphQL code, or adding a DB migration, verify that generated artifacts and registration files are updated.
4. When changing a route, DTO, or exception, verify the client/contract payload, HTTP status, and metric name.

When the environment or a pre-existing issue prevents a test from running, the change description must list the skipped command, reason, and possible impact.
