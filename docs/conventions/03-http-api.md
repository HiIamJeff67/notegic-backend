# HTTP API Conventions

## Gateway routes and middleware

- Browser/client-facing HTTP routes live under
  `runtimes/clientgateway/transports/api/routes`. Its middlewares and interceptors
  are client transport concerns; reusable Gin cookie handlers live in
  `shared/cookies/`. Test-only routes stay in this client transport.
- Routes register URL, middleware, trace operation, metric name, and the
  allowed-permission policy. They do not parse input or call repositories.
- Collection paths use plural kebab-case resource names. Route changes update
  `docs/api-route-design/` in the same change.
- Gateway middleware is the only source of route permission semantics. It
  writes allowed permissions into the trusted Gateway request context; the
  Gateway adapter passes them in the delegation credential.
- Middleware order, trace/metric naming, CORS, CSRF, rate-limit, and auth
  behaviour remain explicit in the route group. Do not hide them in a helper
  router group used once.

## Request and response contracts

- Public and internal operation payloads use `XxxRequestDto` and
  `XxxResponseDto`. `gatewaycontract.Request[Dto]` and
  `gatewaycontract.Response[Dto]`
  carry version, operation, metadata, and the DTO; do not add a nested DTO body
  solely to repeat the envelope.
- `contracts/core/v1/api` package `apicontract` exports the common
  `RequestDto[Header, Body, Param, Query]`.
  composable request shape. Domain request DTOs embed it with anonymous
  concrete struct types directly at the declaration; do not create one-off
  `HeaderDto`, `BodyDto`, `ParamDto`, or `QueryDto` types. The client and Core
  envelopes remain outside these DTOs, so a response never becomes `data.data`.
- A request DTO contains only the needed public header, body, URI, and query
  fields. It never contains actor/context fields, GORM model, DB handle, Gin
  context, or business logic. Trusted actor, permissions, and Core user identity
  are established from delegation and Core middleware, not deserialized from the
  DTO.
- JSON uses `json`, query uses `form`, and URI uses `uri` tags. Validation is
  expressed with `validate` tags and executed at the trust boundary.
- The Gateway `AuthMiddleware` for Core-bound routes only extracts and
  cryptographically parses browser access/refresh tokens. It does not query
  Redis or Core database data. Browser cookies remain a Gateway concern and
  are converted to typed `gatewaycontract.Tokens` in the internal request.
- The Core gateway authentication middleware validates the forwarded token,
  matches its public subject with the delegation subject, and resolves the
  Core-owned internal user identity before invoking a secure endpoint.
- Core-owned user role and plan authorization middleware lives beside Core
  authentication middleware; Gateway retains only client-transport concerns
  such as rate limits, sanitization, CORS, and response interception.
- When access authentication falls back to a valid refresh token, Core rotates
  access and CSRF tokens and places them in the typed internal response
  envelope. The Gateway adapter copies them into its request context; the
  client response interceptor sets the access cookie and exposes only the
  non-sensitive CSRF refresh metadata.
- The Gateway binder binds and validates all public HTTP data. A
  binding/validation failure is mapped to a client-safe Gateway exception
  response; it does not reach a controller or Core service.
- Core service requests are versioned contracts. They contain serializable
  fields only. The calling component (`actor`), optional authenticated public
  user subject (`userSubject`), and allowed permissions come from the verified
  delegation credential rather than an untrusted client field. Use the Core
  `DelegationMiddleware` when a user subject is not required and
  `DelegationAuthenticatedMiddleware` when it is required.

## Controllers and adapters

- A Gateway controller receives a validated request DTO, obtains
  `ctx.Request.Context()`, calls its adapter, and renders a success or safe
  exception response.
- A controller does not bind public HTTP input, query GORM, create a
  transaction, decide ownership, or call a repository directly.
- A Gateway-to-Core adapter lives under
  `runtimes/clientgateway/transports/core/adapters`. It uses its injected
  long-lived client, context propagation, configured timeout, and versioned
  Request/Response mapping. It may be used by REST, GraphQL, and WebSocket.
- Core Gateway endpoints validate the delegation credential and map the internal
  request to a Core service request. They do not invoke a second local adapter;
  their sibling routers only register endpoint method/path pairs.
- Write retry is allowed only for an idempotent operation or one carrying an
  idempotency key. Do not silently retry a general mutation.

## Response, exception, and observability

- Public success remains `{ "success": true, "data": ..., "exception": null }`.
  Public failure remains `{ "success": false, "data": null, "exception": ... }`.
- Gateway owns HTTP exception rendering. Services and repositories return
  service exception or Go error according to their boundary; they never render
  Gin responses.
- Preserve lower-level origin with `WithOrigin(err)` when mapping an expected
  application failure. Never expose internal origin details to the client.
- Every route has a trace operation and metric name. Internal adapter calls
  propagate trace/request identity; do not create a second logging framework.

## Contract change checklist

1. Update the route design document and public/internal Request/Response.
2. Update affected Gateway controller/adapter and Core Gateway endpoint/router/service in the
   same public-operation order.
3. Regenerate artifacts when the contract is generated code input.
4. Verify response shape, status, permission policy, trace/metric name, and
   affected tests.
