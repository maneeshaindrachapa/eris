#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(dirname "$script_dir")

if ! command -v go >/dev/null 2>&1; then
  echo "create-client: go is not installed or not in PATH" >&2
  exit 1
fi

"$script_dir/connect-postgres.sh" --wait

cd "$repo_root"
exec go run ./cmd/client "$@"
