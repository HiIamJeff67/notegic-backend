# ============================== Workspace Commands ============================== #

WORKSPACE_MODULES := contracts shared internal/cli internal/core internal/durablejob internal/email internal/clientgateway internal/apigateway internal/notification internal/realtimegateway test

.PHONY: ci-format ci-vet ci-unit ci-race ci-generated ci-containers staging-deploy staging-smoke kafka-topics \
	compose-integration-up compose-integration-down test-integration test-integration-yjs test-integration-kafka \
	compose-up compose-down test-integration-managed env-check env-encrypt env-decrypt env-edit env-updatekeys env-rotate \
	env-encrypt-all development production test staging \
	test-client-gateway test-api-gateway devlog install-hooks

COMPOSE_INTEGRATION_PROJECT := notegic-integration
COMPOSE_INTEGRATION_FILE := infra/docker/docker-compose.integration.yaml
COMPOSE_FILE ?= docker-compose.yaml
COMPOSE_INIT_FILE ?= docker-compose.init.yaml
COMPOSE_ENCRYPTED_ENV_FILE ?= .env.enc
COMPOSE_SOPS_CONFIG ?= .sops.yaml
SOPS ?= sops
SOPS_CONFIG ?= .sops.yaml
ENVIRONMENTS := development production test staging
ENVIRONMENT ?= development
ENV_DIRECTORY ?= secrets/envs
ENVIRONMENT_GOAL := $(firstword $(filter $(ENVIRONMENTS),$(MAKECMDGOALS)))
SELECTED_ENVIRONMENT := $(if $(ENVIRONMENT_GOAL),$(ENVIRONMENT_GOAL),$(ENVIRONMENT))
ENV_PLAINTEXT_FILE ?= $(if $(filter development,$(SELECTED_ENVIRONMENT)),.env,$(ENV_DIRECTORY)/.env.$(SELECTED_ENVIRONMENT))
ENV_ENCRYPTED_FILE ?= $(if $(filter development,$(SELECTED_ENVIRONMENT)),.env.enc,$(ENV_DIRECTORY)/.env.$(SELECTED_ENVIRONMENT).enc)

env-check:
	@command -v "$(SOPS)" >/dev/null 2>&1 || { echo "sops is required; install SOPS before using env-* targets" >&2; exit 1; }
	@test -f "$(SOPS_CONFIG)" || { echo "missing $(SOPS_CONFIG); create a local SOPS config with your age recipients" >&2; exit 1; }

env-encrypt: env-check
	@test -f "$(ENV_PLAINTEXT_FILE)" || { echo "missing $(ENV_PLAINTEXT_FILE)" >&2; exit 1; }
	@mkdir -p "$$(dirname "$(ENV_ENCRYPTED_FILE)")"
	@temporary_file="$$(mktemp "$(ENV_ENCRYPTED_FILE).tmp.XXXXXX")"; \
	trap 'rm -f "$$temporary_file"' EXIT INT TERM; \
	set -e; \
	umask 077; \
	"$(SOPS)" --config "$(SOPS_CONFIG)" encrypt --input-type dotenv --output-type dotenv "$(ENV_PLAINTEXT_FILE)" > "$$temporary_file"; \
	mv "$$temporary_file" "$(ENV_ENCRYPTED_FILE)"; \
	trap - EXIT INT TERM
	@echo "Encrypted $(ENV_PLAINTEXT_FILE) -> $(ENV_ENCRYPTED_FILE)"

env-encrypt-all: env-check
	@set -eu; \
	for environment in $(ENVIRONMENTS); do \
		echo "Encrypting $$environment"; \
		$(MAKE) --no-print-directory env-encrypt ENVIRONMENT="$$environment"; \
	done

# GNU Make consumes -e itself; the environment goal makes `make env-encrypt -e development` select development.
development production test staging:
	@:

env-decrypt: env-check
	@test -f "$(ENV_ENCRYPTED_FILE)" || { echo "missing $(ENV_ENCRYPTED_FILE)" >&2; exit 1; }
	@mkdir -p "$$(dirname "$(ENV_PLAINTEXT_FILE)")"
	@temporary_file="$$(mktemp "$$(dirname "$(ENV_PLAINTEXT_FILE)")/.notegic-env.XXXXXX")"; \
	trap 'rm -f "$$temporary_file"' EXIT INT TERM; \
	set -e; \
	umask 077; \
	"$(SOPS)" --config "$(SOPS_CONFIG)" decrypt --input-type dotenv --output-type dotenv "$(ENV_ENCRYPTED_FILE)" > "$$temporary_file"; \
	chmod 600 "$$temporary_file"; \
	mv "$$temporary_file" "$(ENV_PLAINTEXT_FILE)"; \
	trap - EXIT INT TERM
	@echo "Decrypted $(ENV_ENCRYPTED_FILE) -> $(ENV_PLAINTEXT_FILE)"

env-edit: env-check
	@test -f "$(ENV_ENCRYPTED_FILE)" || { echo "missing $(ENV_ENCRYPTED_FILE)" >&2; exit 1; }
	@"$(SOPS)" --config "$(SOPS_CONFIG)" edit --input-type dotenv --output-type dotenv "$(ENV_ENCRYPTED_FILE)"

env-updatekeys: env-check
	@test -f "$(ENV_ENCRYPTED_FILE)" || { echo "missing $(ENV_ENCRYPTED_FILE)" >&2; exit 1; }
	@"$(SOPS)" --config "$(SOPS_CONFIG)" updatekeys "$(ENV_ENCRYPTED_FILE)"

env-rotate: env-check
	@test -f "$(ENV_ENCRYPTED_FILE)" || { echo "missing $(ENV_ENCRYPTED_FILE)" >&2; exit 1; }
	@"$(SOPS)" --config "$(SOPS_CONFIG)" rotate --in-place "$(ENV_ENCRYPTED_FILE)"

ci-format:
	@files="$$(find contracts shared internal test -type f -name '*.go' -not -path '*/vendor/*' -print0 | xargs -0 gofmt -l)"; \
	if [ -n "$$files" ]; then echo "Unformatted Go files:"; echo "$$files"; exit 1; fi

ci-vet:
	@set -e; \
	for module in $(WORKSPACE_MODULES); do \
		echo "go vet $$module"; \
		(cd "$$module" && GOWORK=off go vet ./...); \
	done

ci-unit:
	$(MAKE) test-all

ci-race:
	@set -e; \
	for module in $(WORKSPACE_MODULES); do \
		$(MAKE) test-race MODULE="$$module"; \
	done

ci-generated:
	$(MAKE) -C contracts gql-generate
	@git diff --exit-code -- contracts/core/v1/graphql/generated contracts/core/v1/graphql/models

ci-containers:
	@set -e; \
	for runtime in clientgateway apigateway core durablejob email notification realtimegateway yjsworker; do \
		echo "docker build $$runtime"; \
		target=production; \
		if [ "$$runtime" = yjsworker ]; then target=runtime; fi; \
		docker build --target "$$target" --file "internal/$$runtime/Dockerfile" --tag "notegic-ci-$$runtime" .; \
	done

staging-deploy:
	infra/staging/deploy.sh

staging-smoke:
	infra/staging/smoke.sh

compose-up:
	@set -eu; \
	command -v "$(SOPS)" >/dev/null 2>&1 || { echo "sops is required; install SOPS before starting Compose" >&2; exit 1; }; \
	test -f "$(COMPOSE_SOPS_CONFIG)" || { echo "missing $(COMPOSE_SOPS_CONFIG)" >&2; exit 1; }; \
	test -f "$(COMPOSE_ENCRYPTED_ENV_FILE)" || { echo "missing $(COMPOSE_ENCRYPTED_ENV_FILE)" >&2; exit 1; }; \
	temporary_file="$$(mktemp "$${TMPDIR:-/tmp}/notegic-compose-env.XXXXXX")"; \
	trap 'rm -f "$$temporary_file"' EXIT INT TERM; \
	"$(SOPS)" --config "$(COMPOSE_SOPS_CONFIG)" decrypt --input-type dotenv --output-type dotenv "$(COMPOSE_ENCRYPTED_ENV_FILE)" > "$$temporary_file"; \
	chmod 600 "$$temporary_file"; \
	docker compose --project-directory . --env-file "$$temporary_file" --file "$(COMPOSE_FILE)" --file "$(COMPOSE_INIT_FILE)" up --build -d --wait

compose-down:
	@set -eu; \
	command -v "$(SOPS)" >/dev/null 2>&1 || { echo "sops is required; install SOPS before stopping Compose" >&2; exit 1; }; \
	test -f "$(COMPOSE_SOPS_CONFIG)" || { echo "missing $(COMPOSE_SOPS_CONFIG)" >&2; exit 1; }; \
	test -f "$(COMPOSE_ENCRYPTED_ENV_FILE)" || { echo "missing $(COMPOSE_ENCRYPTED_ENV_FILE)" >&2; exit 1; }; \
	temporary_file="$$(mktemp "$${TMPDIR:-/tmp}/notegic-compose-env.XXXXXX")"; \
	trap 'rm -f "$$temporary_file"' EXIT INT TERM; \
	"$(SOPS)" --config "$(COMPOSE_SOPS_CONFIG)" decrypt --input-type dotenv --output-type dotenv "$(COMPOSE_ENCRYPTED_ENV_FILE)" > "$$temporary_file"; \
	chmod 600 "$$temporary_file"; \
	docker compose --project-directory . --env-file "$$temporary_file" --file "$(COMPOSE_FILE)" --file "$(COMPOSE_INIT_FILE)" down --volumes --remove-orphans

CLI_RUN := go -C internal/cli run .

test-all:
	$(CLI_RUN) test-all

test-module:
	@if [ -z "$(MODULE)" ]; then echo "usage: make test-module MODULE=<module>"; exit 1; fi
	$(CLI_RUN) test-module $(MODULE)

test-race:
	@if [ -z "$(MODULE)" ]; then echo "usage: make test-race MODULE=<module>"; exit 1; fi
	$(CLI_RUN) test-race $(MODULE)

test-contracts:
	$(MAKE) -C contracts test

test-shared:
	$(MAKE) -C shared test

test-core:
	$(MAKE) -C internal/core test

test-gateway:
	$(MAKE) -C internal/clientgateway test

test-client-gateway:
	$(MAKE) -C internal/clientgateway test

test-api-gateway:
	$(MAKE) -C internal/apigateway test

devlog:
	./scripts/devlog.sh

install-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks enabled from .githooks"

test-durable-job:
	$(MAKE) -C internal/durablejob test

test-email:
	$(MAKE) -C internal/email test

test-realtime-gateway:
	$(MAKE) -C internal/realtimegateway test

test-notification:
	$(MAKE) -C internal/notification test

test-architecture:
	$(MAKE) -C test test-architecture

compose-integration-up:
	docker compose --project-directory . --project-name $(COMPOSE_INTEGRATION_PROJECT) --file $(COMPOSE_INTEGRATION_FILE) up -d --wait

compose-integration-down:
	docker compose --project-directory . --project-name $(COMPOSE_INTEGRATION_PROJECT) --file $(COMPOSE_INTEGRATION_FILE) down --volumes --remove-orphans

test-integration:
	$(MAKE) -C test test-integration-run

test-integration-yjs:
	@cd internal/yjsworker && NOTEGIC_RUN_INTEGRATION=1 YJS_DB_HOST=127.0.0.1 YJS_DB_PORT=15432 YJS_DB_USER=notegic YJS_DB_PASSWORD=notegic YJS_DB_NAME=notegic_integration pnpm run test:integration

test-integration-kafka:
	$(MAKE) -C test test-integration-kafka-run

test-integration-managed:
	@set -e; \
	trap '$(MAKE) compose-integration-down >/dev/null 2>&1 || true' EXIT; \
	$(MAKE) compose-integration-up; \
	$(MAKE) test-integration; \
	$(MAKE) test-integration-yjs; \
	$(MAKE) test-integration-kafka

test-load-websocket:
	$(MAKE) -C test test-load-websocket

test-soak-websocket:
	$(MAKE) -C test test-soak-websocket

test-load-kafka-lag:
	$(MAKE) -C test test-load-kafka-lag

kafka-topics:
	$(CLI_RUN) kafka topics ensure
