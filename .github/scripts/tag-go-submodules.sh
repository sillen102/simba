#!/usr/bin/env bash
set -euo pipefail

if [ "${1:-}" = "" ]; then
  echo "usage: $0 <version>"
  exit 1
fi

version="$1"
tags=(
  "models/v${version}"
  "telemetry/v${version}"
  "websocket/v${version}"
)

for tag in "${tags[@]}"; do
  if git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; then
    echo "Tag already exists remotely, skipping: ${tag}"
    continue
  fi

  if ! git rev-parse --verify --quiet "refs/tags/${tag}" >/dev/null; then
    git tag "${tag}"
  fi

  git push origin "refs/tags/${tag}"
  echo "Pushed tag: ${tag}"
done
