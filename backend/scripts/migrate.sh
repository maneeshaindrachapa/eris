#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
backend_root=$(dirname "$script_dir")
read_config() { (cd "$backend_root" && go run ./cmd/config "$@"); }
command=${1:-up}
command_arg=${2:-}

set -- $(read_config postgres.container postgres.user postgres.password postgres.database postgres.hostname postgres.port tools.migrate_image)
postgres_container=$1
postgres_user=$2
postgres_password=$3
postgres_db=$4
postgres_hostname=$5
postgres_port=$6
migrate_image=$7

migration_dsn="postgres://${postgres_user}:${postgres_password}@${postgres_hostname}:${postgres_port}/${postgres_db}?sslmode=disable"

if [ "$command" = "create" ]; then
  name=$command_arg
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
  touch "$backend_root/migrations/${version}_${name}.up.sql"
  touch "$backend_root/migrations/${version}_${name}.down.sql"
  echo "Created migrations/${version}_${name}.{up,down}.sql"
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "migrate: docker is not installed or not in PATH" >&2
  exit 1
fi

"$script_dir/connect-postgres.sh" --wait

case "$command" in
  up)
    if [ -n "$command_arg" ]; then set -- up "$command_arg"; else set -- up; fi
    ;;
  down)
    set -- down "${command_arg:-1}"
    ;;
  version)
    set -- version
    ;;
  force)
    if [ -z "$command_arg" ]; then
      echo "Usage: $0 force <version>" >&2
      exit 1
    fi
    set -- force "$command_arg"
    ;;
  goto)
    if [ -z "$command_arg" ]; then
      echo "Usage: $0 goto <version>" >&2
      exit 1
    fi
    set -- goto "$command_arg"
    ;;
  *)
    echo "Usage: $0 {up [N]|down [N]|version|goto VERSION|force VERSION|create NAME}" >&2
    exit 1
    ;;
esac

exec docker run --rm \
  --network "container:$postgres_container" \
  -v "$backend_root/migrations:/migrations:ro" \
  "$migrate_image" \
  -path=/migrations \
  -database "$migration_dsn" \
  "$@"
