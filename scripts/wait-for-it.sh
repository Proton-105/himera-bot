#!/usr/bin/env sh
set -eu

HOST="$1"
PORT="$2"
shift 2 || true

echo "Waiting for ${HOST}:${PORT}..."

while ! nc -z "${HOST}" "${PORT}" >/dev/null 2>&1; do
  echo "  still waiting for ${HOST}:${PORT}..."
  sleep 1
done

echo "Service ${HOST}:${PORT} is up."

if [ "$#" -gt 0 ]; then
  echo "Executing: $*"
  exec "$@"
fi
