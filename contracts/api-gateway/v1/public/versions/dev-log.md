# APIGateway v1 development log

## Current contract baseline

- Published surface: 125 APIGateway operations across nine enabled resource domains.
- Contract format: OpenAPI 3.1.
- Authentication: user-owned `X-API-Key` header; key creation remains on ClientGateway.
- Tooling: Postman 2.1 collection/environment, curl functions, and an HTTP client file.
- Stability: Beta namespace with v1 compatibility rules.

Future entries must identify added, changed, deprecated, and removed operations and link to their migration notes. Never record a breaking change only in application release notes.
