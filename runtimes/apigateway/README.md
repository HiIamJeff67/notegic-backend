# APIGateway

APIGateway is the HTTP entry point for external tools that use Notegic API
keys. It is intentionally separate from the browser-facing ClientGateway.

## Feature contribution

| Capability | APIGateway responsibility |
| --- | --- |
| External integrations | Exposes the public HTTP interface for API-key clients. |
| Workspace automation | Sends validated integration requests to Core, which owns product rules and data. |
| Safe public access | Enforces the integration edge's origin, proxy, and rate-limit policies. |

## Technical design

This Go runtime uses Gin routes, a Core adapter, and Redis-backed rate-limit
state. API-key validation and all business authorization stay with Core, so the
gateway remains a transport boundary rather than a second domain service.

Run the local stack from the repository root with `make compose-up`; it needs
the repository's SOPS-managed environment. Refer to the
[public API documentation](../../docs/integrations/public-api-documentation.md)
before changing the external interface.

## Commands

Run these from this directory:

```sh
make test
make test-race
make test-e2e
make build-api-gateway
```

## Structure

```text
commands/           Runtime command entry point
configs/            APIGateway configuration
data/redis/         Rate-limit persistence
ratelimit/          Rate-limit policy
transports/api/     External HTTP routes and middleware
transports/core/    Core adapter
transports/status/  Health and startup endpoints
test/               Runtime tests
```

See also [public gateway contracts](../../contracts/api-gateway/v1/public/README.md)
and the [root runtime map](../../README.md).
