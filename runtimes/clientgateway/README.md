# ClientGateway

ClientGateway is the HTTP and GraphQL entry point used by Notegic applications.
It turns browser-facing requests into authenticated calls to the runtimes that
own the underlying work.

## Feature contribution

| User-facing capability | ClientGateway responsibility |
| --- | --- |
| Sign-in and account access | Issues and clears client session cookies, then forwards account requests to Core. |
| Workspace and routine actions | Accepts client API and GraphQL requests and delegates product decisions to Core. |
| Notifications | Proxies the authenticated notification experience to Notification. |
| Safe browser access | Applies origin, proxy, cookie, and request-rate protections at the client edge. |

## Technical design

This Go runtime uses Gin for HTTP routing. It keeps client-facing cookie and
rate-limit behavior at the edge, uses Redis for rate-limit state, and calls
Core and Notification through runtime adapters. It does not own product data or
make authorization decisions that belong to Core.

Start the full local stack from the repository root with `make compose-up`.
It loads the encrypted environment through SOPS; use the root
[environment-secrets runbook](../../docs/runbooks/environment-secrets.md)
rather than committing a local `.env`.

## Commands

Run these from this directory:

```sh
make test
make test-race
make test-e2e
make build-client-gateway
```

## Structure

```text
commands/           Runtime command entry point
configs/            ClientGateway configuration
data/redis/         Rate-limit persistence
ratelimit/          Rate-limit policy
transports/api/     Public routes and middleware
transports/core/    Core adapter
transports/notification/ Notification adapter
transports/status/  Health and startup endpoints
test/               Runtime tests
```

See also [public gateway contracts](../../contracts/client-gateway/v1/public/README.md)
and the [root runtime map](../../README.md).
