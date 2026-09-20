#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(dirname "$script_dir")
env_file="$repo_root/.env"

if [ -f "$env_file" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$env_file"
  set +a
fi

postgres_container=${POSTGRES_CONTAINER:-eris-postgres}
postgres_user=${POSTGRES_USER:-eris}
postgres_password=${POSTGRES_PASSWORD:-eris}
postgres_db=${POSTGRES_DB:-eris}
postgres_hostname=${POSTGRES_HOSTNAME:-localhost}
postgres_port=${POSTGRES_PORT:-5432}
migrate_image=${MIGRATE_IMAGE:-migrate/migrate:v4.18.3}

migration_dsn="postgres://${postgres_user}:${postgres_password}@${postgres_hostname}:${postgres_port}/${postgres_db}?sslmode=disable"
command=${1:-up}

if [ "$command" = "create" ]; then
  name=${2:-}
  if [ -z "$name" ]; then
    echo "Usage: $0 create <migration_name>" >&2
    exit 1
  fi
  case "$name" in
    *[!a-z0-9_]*)
      echo "migrate: migration name may contain lowercase letters, numbers, and underscores only" >&2
      exit 1
      ;;
  esac

  version=$(date -u +%Y%m%d%H%M%S)
  touch "$repo_root/migrations/${version}_${name}.up.sql"
  touch "$repo_root/migrations/${version}_${name}.down.sql"
  echo "Created migrations/${version}_${name}.{up,down}.sql"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "migrate: docker is not installed or not in PATH" >&2
  exit 1
fi

"$script_dir/connect-postgres.sh" --wait

shift || true
case "$command" in
  up)
    set -- up "$@"
    ;;
  down)
    set -- down "${1:-1}"
    ;;
  version)
    set -- version
    ;;
  force)
    if [ "$#" -ne 1 ]; then
      echo "Usage: $0 force <version>" >&2
      exit 1
    fi
    set -- force "$1"
    ;;
  goto)
    if [ "$#" -ne 1 ]; then
      echo "Usage: $0 goto <version>" >&2
      exit 1
    fi
    set -- goto "$1"
    ;;
  *)
    echo "Usage: $0 {up [N]|down [N]|version|goto VERSION|force VERSION|create NAME}" >&2
    exit 1
    ;;
esac

exec docker run --rm \
  --network "container:$postgres_container" \
  -v "$repo_root/migrations:/migrations:ro" \
  "$migrate_image" \
  -path=/migrations \
  -database "$migration_dsn" \
  "$@"
