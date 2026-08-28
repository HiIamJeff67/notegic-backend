#!/usr/bin/env sh

set -eu

: "${IMAGE_TAG:?IMAGE_TAG is required}"
: "${IMAGE_REGISTRY:?IMAGE_REGISTRY is required}"

IMAGE_REGISTRY=$(printf '%s' "$IMAGE_REGISTRY" | tr '[:upper:]' '[:lower:]')
export IMAGE_REGISTRY

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

export CLIENT_GATEWAY_IMAGE="$IMAGE_REGISTRY/notegic-client-gateway:$IMAGE_TAG"
export API_GATEWAY_IMAGE="$IMAGE_REGISTRY/notegic-api-gateway:$IMAGE_TAG"
export CORE_IMAGE="$IMAGE_REGISTRY/notegic-core:$IMAGE_TAG"
export DURABLE_JOB_IMAGE="$IMAGE_REGISTRY/notegic-durablejob:$IMAGE_TAG"
export NOTIFICATION_IMAGE="$IMAGE_REGISTRY/notegic-notification:$IMAGE_TAG"
export EMAIL_IMAGE="$IMAGE_REGISTRY/notegic-email:$IMAGE_TAG"
export REALTIME_GATEWAY_IMAGE="$IMAGE_REGISTRY/notegic-realtimegateway:$IMAGE_TAG"
export YJS_WORKER_IMAGE="$IMAGE_REGISTRY/notegic-yjsworker:$IMAGE_TAG"

docker compose --project-directory "$root_directory" --env-file "$compose_env_file" -f "$compose_file" -f "$compose_init_file" pull
docker compose --project-directory "$root_directory" --env-file "$compose_env_file" -f "$compose_file" -f "$compose_init_file" up -d --no-build --remove-orphans
