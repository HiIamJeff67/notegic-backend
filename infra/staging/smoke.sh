#!/usr/bin/env sh

set -eu

root_directory=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
compose_file=${COMPOSE_FILE:-"$root_directory/infra/docker/docker-compose.prod.yaml"}
compose_init_file=${COMPOSE_INIT_FILE:-"$root_directory/infra/docker/docker-compose.init.prod.yaml"}
compose_env_file=${COMPOSE_ENV_FILE:-/etc/notegic/staging.env}
compose_encrypted_env_file=${COMPOSE_ENCRYPTED_ENV_FILE:-}
temporary_env_file=

cleanup() {
	if [ -n "$temporary_env_file" ]; then
		rm -f "$temporary_env_file"
	fi
}

trap cleanup EXIT INT TERM

if [ -n "$compose_encrypted_env_file" ]; then
	command -v sops >/dev/null 2>&1 || { echo "sops is required when COMPOSE_ENCRYPTED_ENV_FILE is set" >&2; exit 1; }
	sops_config_file=${SOPS_CONFIG_FILE:-"$root_directory/.sops.yaml"}
	test -f "$sops_config_file" || { echo "missing SOPS config: $sops_config_file" >&2; exit 1; }
	: "${SOPS_AGE_KEY_FILE:?SOPS_AGE_KEY_FILE is required when COMPOSE_ENCRYPTED_ENV_FILE is set}"
	test -f "$SOPS_AGE_KEY_FILE" || { echo "missing SOPS age key file: $SOPS_AGE_KEY_FILE" >&2; exit 1; }
	export SOPS_AGE_KEY_FILE
	temporary_env_file=$(mktemp "${TMPDIR:-/tmp}/notegic-staging-env.XXXXXX")
	sops --config "$sops_config_file" decrypt \
		--input-type dotenv \
		--output-type dotenv \
		"$compose_encrypted_env_file" > "$temporary_env_file"
	chmod 600 "$temporary_env_file"
	compose_env_file=$temporary_env_file
fi

compose() {
	docker compose --project-directory "$root_directory" --env-file "$compose_env_file" -f "$compose_file" -f "$compose_init_file" "$@"
}

check_http() {
	service=$1
	port=$2
	path=$3
	maximum_attempts=${SMOKE_MAXIMUM_ATTEMPTS:-30}
	attempt=1

	while [ "$attempt" -le "$maximum_attempts" ]; do
		if compose exec -T "$service" wget -q -O /dev/null "http://127.0.0.1:${port}${path}"; then
			echo "PASS ${service}${path}"
			return 0
		fi

		sleep "${SMOKE_RETRY_DELAY_SECONDS:-2}"
		attempt=$((attempt + 1))
	done

	echo "FAIL ${service}${path}" >&2
	return 1
}

check_yjs_worker() {
	maximum_attempts=${SMOKE_MAXIMUM_ATTEMPTS:-30}
	attempt=1

	while [ "$attempt" -le "$maximum_attempts" ]; do
		if compose exec -T notegic-yjs-worker node -e \
			"fetch('http://127.0.0.1:' + process.env.YJS_WORKER_PORT + '/healthz').then(response => process.exit(response.ok ? 0 : 1)).catch(() => process.exit(1))"; then
			echo "PASS notegic-yjs-worker/healthz"
			return 0
		fi

		sleep "${SMOKE_RETRY_DELAY_SECONDS:-2}"
		attempt=$((attempt + 1))
	done

	echo "FAIL notegic-yjs-worker/healthz" >&2
	return 1
}

check_http notegic-client-gateway "${DOCKER_GIN_PORT:-7777}" /startedz
check_http notegic-client-gateway "${DOCKER_GIN_PORT:-7777}" /healthz
check_http notegic-api-gateway "${DOCKER_API_GATEWAY_PORT:-7780}" /startedz
check_http notegic-api-gateway "${DOCKER_API_GATEWAY_PORT:-7780}" /healthz
check_http notegic-core 7778 /startedz
check_http notegic-core 7778 /healthz
check_http notegic-realtime-gateway 7779 /startedz
check_http notegic-realtime-gateway 7779 /healthz
check_http notegic-durable-job 8082 /startedz
check_http notegic-durable-job 8082 /healthz
check_http notegic-notification 7781 /startedz
check_http notegic-notification 7781 /healthz
check_http notegic-email 8081 /startedz
check_http notegic-email 8081 /healthz
check_yjs_worker
