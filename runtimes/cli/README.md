# Backend CLI

The Backend CLI is the repository-level command runner for common maintenance
and verification tasks. It is not a user-facing product runtime.

## Responsibilities

- Runs the shared test and race-test workflow across Go modules.
- Provides repository maintenance commands, including Kafka topic setup.
- Gives the root Makefile one consistent command surface for local work and
  continuous integration.

## Technical design

The CLI is a small Go program using Cobra. It coordinates existing runtime and
shared-module commands; it must not contain product rules or duplicate a
runtime's application logic. Add a command here only when it is genuinely
repository-wide.

## Commands

From the repository root:

```sh
go -C runtimes/cli run . --help
make test-all
make kafka-topics
```

The root Makefile also exposes the CI checks used by automated pipelines:

```sh
make ci-format
make ci-vet
make ci-unit
make ci-race
make ci-generated
make ci-containers
```

## Structure

```text
main.go                  CLI entry point
root_command.go          Root Cobra command
test_command.go          Shared test-command orchestration
kafka_topics_command.go  Kafka topic setup command
```

Use the [root README](../../README.md) to find a product runtime; use this CLI
only for repository-wide operations.
