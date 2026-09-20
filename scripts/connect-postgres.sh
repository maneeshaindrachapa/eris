#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(dirname "$script_dir")
env_file="$repo_root/.env"

if ! command -v docker >/dev/null 2>&1; then
  echo "connect-postgres: docker is not installed or not in PATH" >&2
  exit 1
fi

if [ -f "$env_file" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$env_file"
  set +a
fi

postgres_container=${POSTGRES_CONTAINER:-eris-postgres}
postgres_image=${POSTGRES_IMAGE:-postgres:17}
postgres_user=${POSTGRES_USER:-eris}
postgres_password=${POSTGRES_PASSWORD:-eris}
postgres_db=${POSTGRES_DB:-eris}
postgres_port=${POSTGRES_PORT:-5432}
postgres_volume=${POSTGRES_VOLUME:-eris-postgres-data}

if ! docker info >/dev/null 2>&1; then
  echo "connect-postgres: docker is not running" >&2
  exit 1
fi

if docker container inspect "$postgres_container" >/dev/null 2>&1; then
  if [ "$(docker inspect -f '{{.State.Running}}' "$postgres_container")" != "true" ]; then
    echo "Starting PostgreSQL container $postgres_container..."
    docker start "$postgres_container" >/dev/null
  fi
else
  echo "Creating PostgreSQL container $postgres_container from $postgres_image..."
  docker run --name "$postgres_container" \
    -e POSTGRES_USER="$postgres_user" \
    -e POSTGRES_PASSWORD="$postgres_password" \
    -e POSTGRES_DB="$postgres_db" \
    -p "$postgres_port:5432" \
    -v "$postgres_volume:/var/lib/postgresql/data" \
    -d "$postgres_image" >/dev/null
fi

echo "Waiting for PostgreSQL to accept connections..."
attempt=0
until docker exec "$postgres_container" pg_isready -U "$postgres_user" -d "$postgres_db" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo "connect-postgres: PostgreSQL did not become ready" >&2
    docker logs --tail 20 "$postgres_container" >&2
    exit 1
  fi
  sleep 1
done

if [ "${1:-}" = "--init" ]; then
  exec "$script_dir/migrate.sh" up
fi

if [ "${1:-}" = "--wait" ]; then
  exit 0
fi

if [ -t 0 ] && [ -t 1 ]; then
  exec docker exec -it "$postgres_container" psql -U "$postgres_user" -d "$postgres_db" "$@"
fi

exec docker exec -i "$postgres_container" psql -U "$postgres_user" -d "$postgres_db" "$@"
