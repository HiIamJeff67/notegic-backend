# Notegic APIGateway v1 public API

This directory contains the machine-readable and human-readable contract for all 125 versioned routes currently exposed by APIGateway v1.

The published domains are RootShelf, SubShelf, Material, BlockPack, Block, Station, Routine, RoutineTask, and RoutineTag. Client-only auth, user/account, notification, realtime, GraphQL, and static routes are intentionally excluded.

- **Canonical contract:** `openapi/openapi.json` (OpenAPI 3.1)
- **Rules:** `rules/`
- **Endpoint catalog:** `reference/endpoints.md`
- **Runnable examples:** `examples/curl/all-endpoints.sh` and `examples/http/all-endpoints.http`
- **API key example:** send `X-API-Key` using the value in your private environment.
- **Postman:** import both JSON files in `postman/`
- **Version records:** `versions/dev-log.md` and `versions/comparison.md`

The generated artifacts are refreshed from routes and Go DTOs with:

```bash
make -C contracts public-api-gen
```

No real account, password, cookie, CSRF token, or API secret belongs in this directory.
