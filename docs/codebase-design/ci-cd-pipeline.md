# CI/CD and Jenkins Pipeline

NOT-43 defines the division of responsibilities between GitHub Actions and
Jenkins. Both may invoke only commands already exposed by the root `Makefile`
or `internal/cli`; a pipeline must not duplicate a second test workflow.

## GitHub Actions

`.github/workflows/ci.yml` is the quality gate for every pull request, pushes to
`main`/`refactor/**`, and version tags. It includes:

- formatting, `go vet`, unit tests, and `go test -race` for every Go module;
- a clean-tree check after regenerating GraphQL artifacts;
- production container builds for the six Go runtimes and the Yjs Worker;
- publishing runtime images to GitHub Container Registry after a version tag
  passes every gate.

`.github/workflows/integration.yml` runs Docker-backed integration tests only
when manually triggered or scheduled. It first starts the PostgreSQL, Redis,
and Kafka services defined by
`infra/docker/docker-compose.integration.yaml` through the root
`Makefile`'s `compose-integration-up`, runs root integration and
runtime Kafka integration tests, and finally cleans up with
`compose-integration-down`. Local development can use the same root targets;
when Compose is already running, `make test-integration` or
`make test-integration-kafka` runs tests without taking over container
management.

`.github/workflows/staging.yml` is a promotion workflow requiring approval from
the `staging` environment. On a self-hosted runner labeled `staging`, it merges
`infra/docker/docker-compose.prod.yaml` with
`infra/docker/docker-compose.init.prod.yaml` using the selected GHCR tag,
promotes immutable images through `infra/staging/deploy.sh`, and checks every runtime's
`/startedz` and `/healthz` with `infra/staging/smoke.sh`. Compose accepts
`CLIENT_GATEWAY_IMAGE`, `API_GATEWAY_IMAGE`, `CORE_IMAGE`, `DURABLE_JOB_IMAGE`, `NOTIFICATION_IMAGE`, `EMAIL_IMAGE`,
`REALTIME_GATEWAY_IMAGE`, and `YJS_WORKER_IMAGE`, so promotion does not rebuild
images. The staging runner may provide plaintext settings in
`/etc/notegic/staging.env` for compatibility, or provide an encrypted
`/etc/notegic/staging.env.enc` plus an age identity and set
`COMPOSE_ENCRYPTED_ENV_FILE`, `SOPS_CONFIG_FILE`, and `SOPS_AGE_KEY_FILE`.
The deployment and smoke scripts verify the private key path, decrypt the
latter into a temporary file, and remove it on exit. Secrets must not be
committed or uploaded as artifacts. Compose logs after deployment are uploaded
as an artifact with 14-day retention.

## Jenkins

The root `Jenkinsfile` is the delivery pipeline for a self-hosted agent:

1. checkout, format, vet, unit, race, and generated-contract gates;
2. an optional production container build;
3. promotion of immutable images and the same smoke script on an agent labeled
   `notegic-staging`, controlled by the `DEPLOY_STAGING` and `IMAGE_TAG`
   parameters;
4. optional integration tests controlled by `RUN_INTEGRATION`. Tests share
   `infra/docker/docker-compose.integration.yaml` with GitHub Actions, and the
   Compose stack is cleaned up afterward.

The Jenkins staging agent must provide Docker, the Compose plugin, Git, and a
Jenkins credential that can access GHCR. Staging deployments may promote only
immutable image tags that have passed CI; the staging agent must not recompile
source.

Jenkins does not replace the GitHub Actions PR gate, and this issue does not
perform a production rollout, migration rollback, or disaster recovery.
Image publishing, environment secrets, and deployment permissions must be
managed by Jenkins credentials or GitHub OIDC and must not be committed to the
repository.

## Local Equivalent Commands

```sh
make ci-format
make ci-vet
make ci-unit
make ci-race
make ci-generated
make ci-containers
```

The staging runner's delivery commands are:

```sh
IMAGE_REGISTRY=ghcr.io/ORG/REPO IMAGE_TAG=TAG \
COMPOSE_ENCRYPTED_ENV_FILE=/etc/notegic/secrets/envs/.env.staging.enc \
SOPS_CONFIG_FILE=/etc/notegic/sops/.sops.yaml \
SOPS_AGE_KEY_FILE=/etc/notegic/sops/age/keys-staging.txt \
make staging-deploy

COMPOSE_ENCRYPTED_ENV_FILE=/etc/notegic/secrets/envs/.env.staging.enc \
SOPS_CONFIG_FILE=/etc/notegic/sops/.sops.yaml \
SOPS_AGE_KEY_FILE=/etc/notegic/sops/age/keys-staging.txt \
make staging-smoke
```

Integration tests use `make test-integration` and
`make test-integration-kafka`; these commands only execute tests and require
Docker-backed dependencies to be available. Use
`make test-integration-managed` for the complete lifecycle, or run
`compose-integration-up`, the test commands, and `compose-integration-down`
separately.
