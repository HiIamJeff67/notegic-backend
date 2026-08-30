# Exception Conventions

## Ownership and envelope

- `contracts/types/exceptions` is a pure application envelope. It may use the standard
  library, but must not import `shared`, Gateway, a microservice,
  platform observability, Gin, GORM, or generated GraphQL code.
- `contracts/types/exceptions` creates the shared envelope with `exceptions.New()`.
  Its optional final `isInternal ...bool` accepts only the first value and
  cannot be changed after construction. It must not own a domain factory or a
  numeric code registry.
- Gateway and each microservice may define a local factory in their own
  `exceptions/` package when two or more callers share the same domain or
  operational error semantics. Core domain factories remain in
  `runtimes/core/exceptions/`; worker, cache, renderer, delivery and
  notification factories are runtime-local helpers. A local helper is never
  imported across a Gateway/service boundary.
- Runtime-local exception packages follow the Core shape: one file per owned
  domain, an `exception.go` base helper type, and categorized operation files
  when the domain has multiple concerns. Each package exposes an exported
  runtime-specific helper type and a `New<Domain>Exception()` factory. The
  package-local base is created with `NewException(domain)`. Named helper methods such as
  `PayloadDecodeFailed` or `InvalidPayload` must return
  `*contracts/types/exceptions.Exception`; the runtime-specific type is never
  the service or transport return type. Do not create a package-level domain
  instance, expose a generic `New(reason, ...)` factory, or add a catch-all
  `errors.go`.
- Runtime-wide exception packages may be used by that runtime's services,
  workers, and handlers. If an exception is used by only one component, keep
  it at that component boundary instead of promoting it into a shared or
  repository exception package.
- Every exception implementation file has its own matching unit-test file:
  `renderer_exception.go` is tested by `renderer_exception_test.go`, and so on.
- Core's `runtimes/core/exceptions/exception.go` and DurableJob's equivalent
  define their own runtime-local `Exception` helper, which composes the
  contract `exceptions.Exception` and stores the domain. PostgreSQL repositories
  use the separate `RepositoryException` under
  `shared/platform/postgres/repositories/exceptions/`; runtime exception
  packages must not import it. Each `*_exception.go` defines an exported domain
  exception type and a `New<Domain>Exception()` factory. Runtimes must not
  expose global domain values such as `Auth` or `Shelf`.
- Runtime-local services, repositories and workers return ordinary `error` or
  the shared `*contracts/types/exceptions.Exception` produced by their local
  helper. They must not return the runtime-specific helper type itself. The
  helper may carry HTTP status only because the shared exception envelope owns
  that transport metadata; it must not format Gin responses or expose public
  response semantics. One-off errors, including generic utility failures, use
  ordinary `errors.New`/`fmt.Errorf` at the call site.
- Set `Reason`, `Domain`, and `Operation` explicitly. They are stable,
  machine-readable application properties; `Message` is the human-readable
  explanation. Use `WithOrigin(err)` for the underlying Go error and
  `WithDetails()` only for in-process diagnostics.
- `Exception` never imports observability and has no `Log()` method. At an
  owning runtime boundary, use `logs.NotegicLogger.JSON(ctx, slog.LevelError,
  exception.String(), exception)` to serialize and record the exception. The
  logger accepts `any`, owns JSON marshaling, and returns its marshal error;
  callers that can safely continue may explicitly discard that logging error.
  `Exception.String()` is diagnostic-only and may contain origin/details; it
  must never be sent to a client.
- Root `ExceptionCode`, `ExceptionPrefix`, generic domain helpers, and domain
  factories do not exist. Domain factories are local to their owning runtime
  or component; callers must not recreate a root compatibility layer.

```go
exception := exceptions.New(
	"NotFound",
	"RootShelf",
	"GetMyRootShelfById",
	"Root shelf was not found",
	http.StatusNotFound,
)

exception := exceptions.New(
	"DatabaseFailure",
	"API",
	"CreateRootShelf",
	"Failed to persist the root shelf",
	http.StatusInternalServerError,
	true,
).WithOrigin(result.Error)
```

## Public safety boundary

- Only `shared/util/responsewriter.ToPublic()` may convert an exception
  into a browser-facing exception. It records the original internal exception
  before conversion and removes `Origin`, `Details`, numeric compatibility
  codes, and all implementation diagnostics.
- An internal exception becomes the safe, generic `InternalServerError` unless
  the owner explicitly provides an approved fallback with
  `WithPublicFallback()`. A fallback may expose only a safe reason, domain,
  operation, message, HTTP status, and retryability.
- Gateway REST and GraphQL response paths must use the response writer. Do not
  construct an exception JSON body directly and do not pass `Origin.Error()` to
  a response or GraphQL extension.
- Gateway records the original exception through the platform logger before
  calling `ToPublic()`. Internal service transport responses carry the same
  safe `reason`, `domain`, `operation`, `message`, and `retryable` fields; a
  Gateway adapter reconstructs that envelope without domain-specific mapping.
- `ToPublic()` is idempotent: calling it for an already public exception keeps
  its safe fields and still drops diagnostics.

```go
exception := exceptions.New(
	"StorageFailure",
	"API",
	"UploadMaterial",
	"storage provider returned an internal response",
	http.StatusInternalServerError,
	true,
).WithPublicFallback(exceptions.PublicFallback{
	Reason:         "ServiceUnavailable",
	Domain:         "Storage",
	Operation:      "UploadMaterial",
	Message:        "Storage is temporarily unavailable",
	HTTPStatusCode: http.StatusServiceUnavailable,
	Retryable:      true,
})
```

## Context ownership

- `shared/contexts` contains only generic `context.Context` value
  helpers; it never imports exceptions or framework code.
- `runtimes/clientgateway/contexts` owns Gin/request-context parsing and the
  route-declared permission set used to construct an outbound delegation.
- `runtimes/core/contexts` owns verified API-service context values,
  including permissions reconstructed from the verified delegation credential.
  Core services must not trust client DTO fields for actor identity or route
  permission policy.

## `exceptions.Cover()`

`exceptions.Cover(existing, fallbacks)` is kept for a single operation with two
or more alternative failure conditions. It returns `existing` unchanged, or the
first satisfied fallback when `existing` is nil. Put an underlying result error
before empty-result and invariant fallbacks so the diagnostic cause is retained.

Use a direct `if` for one clear condition. `Cover()` does not log, roll back, or
write a response; return or finish transaction handling immediately after it
selects an exception.
