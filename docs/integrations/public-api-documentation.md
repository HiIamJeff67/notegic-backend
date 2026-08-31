# Public API documentation decision

## Result

The external integration API contract belongs to APIGateway. Each runtime also owns a public, runtime-specific contract:

The generated APIGateway contract contains all 125 currently emitted operations:

- `contracts/api-gateway/v1/public/` is the only externally advertised v1 contract.
- `contracts/client-gateway/v1/public/` documents the ClientGateway user/client boundary.
- `contracts/realtime-gateway/v1/public/` documents RealtimeGateway HTTP and WebSocket output for realtime clients.

APIGateway REST uses OpenAPI 3.1 for the nine enabled resource domains: RootShelf, SubShelf, Material, BlockPack, Block, Station, Routine, RoutineTask, and RoutineTag. RealtimeGateway HTTP uses OpenAPI 3.1 and its multiplex WebSocket messages use AsyncAPI 3.0. Postman 2.1 collections and environments are committed as importable derived artifacts.

## Content boundary

These directories contain only API contracts, operational rules, complete request examples, and test-tool imports. Product usage teaching and UI workflows remain separate frontend concerns. `rules/` describes constraints; it is not a tutorial or quick-start page.

The human-readable directories are named `rules/` because they describe constraints: API-key authentication, CORS, rate limits, retry behavior, version compatibility, and stable errors.

Each contract also owns `versions/dev-log.md` and `versions/comparison.md`. The current comparison records the single published v1 baseline and defines the fields that must be compared when another version is introduced, without inventing unreleased behavior.

## Authentication decision

ClientGateway continues to use access/refresh JWT cookies for the existing web/client flow. APIGateway v1 uses a user-owned API key in `X-API-Key`; a consumer creates and revokes keys through the authenticated client surface, then sends the secret to APIGateway. The server stores only a SHA-256 digest and never returns the secret after creation.

The APIGateway key is verified at the edge and again in Core through the API key cache/DB fallback. Core writes the same actor fields used by client requests into context, but does not run browser `AuthMiddleware()`. Unauthorized rate limiting remains IP-first; API key ID is only an auxiliary dimension.

This mechanism is suitable for a user's own client and trusted server-side integrations. Third-party browser apps must not collect Notegic credentials and also require an explicitly allowed Origin.

## Maintenance

Routes and Go DTO contracts remain authoritative. Run:

```bash
make -C contracts public-api-gen
```

The generator refreshes both OpenAPI documents, AsyncAPI, Postman files, examples, indexes, rules, and this decision record. CI should later run the generator and reject a dirty diff so route changes cannot silently bypass public documentation.

No generated example contains a real credential. Exported developer environments containing passwords, cookies, CSRF values, OAuth codes, or realtime tickets must never be committed.
