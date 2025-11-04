#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "[reset-db] Stopping stack and dropping volumes..."
docker compose down -v

echo "[reset-db] Starting Postgres and Redis..."
docker compose up -d db redis

echo "[reset-db] Waiting for Postgres on localhost:5432..."
"${ROOT_DIR}/scripts/wait-for-it.sh" localhost 5432

echo "[reset-db] Starting app (will run migrations on startup if wired)..."
docker compose up -d bot

echo "[reset-db] Done."
