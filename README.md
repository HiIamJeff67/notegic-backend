<a><img src="global/images/logo/header-image.png" alt="Notegic" /></a>

# Notegic Backend

Notegic is a workspace for organizing knowledge, editing shared documents, and
planning recurring work. This repository provides the server-side capabilities
that support the Notegic applications and external integrations.

## Highlights

- **Knowledge workspace:** supports shelves, materials, block packs, sharing,
  and permissions.
- **Document collaboration:** keeps shared block-pack documents available and
  synchronized for active collaborators.
- **Routine planning:** supports stations, routines, tasks, dependencies, and
  task history.
- **Accounts and access:** handles sign-in, security workflows, API keys, and
  account-level settings.
- **Updates and delivery:** delivers in-app notifications, email, and live
  application updates.
- **Integration surface:** provides a client-facing API and a separate API-key
  entry point for external tools.

## Runtimes

Each runtime has its own README with responsibilities, technical design, local
commands, configuration notes, and internal folder structure.

| Runtime | Role |
| --- | --- |
| [ClientGateway](runtimes/clientgateway/README.md) | Entry point for Notegic applications. |
| [APIGateway](runtimes/apigateway/README.md) | Entry point for external API-key integrations. |
| [Core](runtimes/core/README.md) | Product rules, permissions, and durable workspace data. |
| [RealtimeGateway](runtimes/realtimegateway/README.md) | Live connections, presence, and realtime delivery. |
| [YjsWorker](runtimes/yjsworker/README.md) | Shared-document persistence and maintenance. |
| [DurableJob](runtimes/durablejob/README.md) | Scheduled and durable routine-task work. |
| [Notification](runtimes/notification/README.md) | Persistent in-app notifications. |
| [Email](runtimes/email/README.md) | Transactional email delivery. |
| [CLI](runtimes/cli/README.md) | Repository maintenance and verification commands. |

## Repository structure

```text
contracts/          Versioned agreements between runtimes and clients
runtimes/           Independently runnable backend applications
  clientgateway/    Client application entry point
  apigateway/       External integration entry point
  core/             Workspace and account domain runtime
  realtimegateway/  Live-update and collaboration gateway
  yjsworker/        Shared-document worker
  durablejob/       Durable task worker
  notification/     In-app notification runtime
  email/            Email delivery runtime
  cli/              Repository command runner
shared/             Cross-runtime platform code and utilities
docs/               Architecture, conventions, and operational guides
infra/              Local and deployed infrastructure definitions
test/               Cross-runtime verification
```

See the README inside a runtime for its implementation details. Repository-wide
architecture and operational documentation lives in [docs/](docs/).

## Licensing

Project code is distributed under the Notegic proprietary license:

- `LICENSE.md` — English
- `LICENSE(tw).md` — Traditional Chinese

Third-party notices and license texts are preserved under `LICENSES/`.

<!-- DEVLOG:START -->
## Development log

This section is automatically maintained from recent Git history. Detailed intent belongs in commit messages and design documents.

### Recent snapshots

- [2026-09/2026-09-05](docs/devlogs/2026-09/2026-09-05.md)
- [2026-09/2026-09-04](docs/devlogs/2026-09/2026-09-04.md)
- [2026-09/2026-09-03](docs/devlogs/2026-09/2026-09-03.md)
- [2026-09/2026-09-02](docs/devlogs/2026-09/2026-09-02.md)
- [2026-09/2026-09-01](docs/devlogs/2026-09/2026-09-01.md)
<!-- DEVLOG:END -->
