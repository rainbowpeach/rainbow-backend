#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

if [[ -f ".env" ]]; then
  echo "using .env from ${ROOT_DIR}"
else
  echo ".env not found, falling back to process environment"
fi

GOCACHE=/tmp/go-build-cache \
GOMODCACHE=/tmp/go-mod-cache \
go run ./cmd/server
